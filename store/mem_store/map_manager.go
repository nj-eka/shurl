package mem_store

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"
	"time"
)

var ErrClosed = errors.New("closed")
var ErrNotFound = errors.New("not found")
var ErrInvalidValue = errors.New("invalid value")

type mapLinkManager struct {
	path         string
	stop         <-chan struct{}
	completed    chan struct{}
	err          error
	mapLinks     map[string]*Link // int -> string for json marshaling
	mapIndexUrls map[string]string
	chOps        chan request
	wg           sync.WaitGroup
	next         int
}

type request map[string]interface{}

type response struct {
	value interface{}
	err   error
}

type addedResult struct {
	id    int
	added bool
}

func newMapManager(stop <-chan struct{}, path string) (*mapLinkManager, error) {
	mapLinks := make(map[string]*Link)
	mapIndexUrls := make(map[string]string)
	next := 0
	if path != "" {
		if file, err := os.OpenFile(path, os.O_RDONLY, 0); err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		} else {
			defer file.Close()
			if err := json.NewDecoder(file).Decode(&mapLinks); err != nil {
				return nil, err
			}
			mapIndexUrls = make(map[string]string, len(mapLinks))
			for cid, link := range mapLinks {
				mapIndexUrls[link.TargetUrl] = cid
				if next < link.Id {
					next = link.Id
				}
			}
		}
	}
	ms := mapLinkManager{
		path:         path,
		stop:         stop,
		mapLinks:     mapLinks,
		mapIndexUrls: mapIndexUrls,
		next:         next,
		// buffer length doesn't matter here in fact cuz blocking will be in any case, whether it is writing or reading
		// operations are serialized / linearized as an alternative to mutex, but with the possibility of unified logging of operations
		chOps:     make(chan request),
		completed: make(chan struct{}),
		wg:        sync.WaitGroup{},
	}
	ms.startProcessOperations()
	return &ms, nil
}

func (mlm *mapLinkManager) startProcessOperations() {
	go func() {
		<-mlm.stop
		mlm.stop = nil // == don't accept new operation (+ panic on re-closing)
		go func() {
			mlm.wg.Wait() // wait for last operation being completed
			close(mlm.chOps)
			mlm.save()
		}()
	}()
	go func() {
		defer close(mlm.completed)
		for request := range mlm.chOps {
			op := request["op"]
			switch {
			case op == "addLink":
				resCh := request["rc"].(chan response)
				url := request["url"].(string)
				var link *Link
				var sid string
				var ok bool
				if sid, ok = mlm.mapIndexUrls[url]; !ok {
					mlm.next++
					sid = strconv.Itoa(mlm.next)
					link = &Link{
						Id:        mlm.next,
						TargetUrl: url,
						CreatedAt: time.Now().UTC(),
						DeletedAt: nil,
						ExpiredAt: request["expiredAt"].(*time.Time),
						Hits:      0,
					}
					mlm.mapLinks[sid] = link
					mlm.mapIndexUrls[url] = sid
				}
				link = mlm.mapLinks[sid]
				resCh <- response{value: &addedResult{id: link.Id, added: !ok}}
			case op == "getLink":
				sid := strconv.Itoa(request["id"].(int))
				resCh := request["rc"].(chan response)
				if link, ok := mlm.mapLinks[sid]; ok {
					resCh <- response{value: link}
				} else {
					resCh <- response{err: ErrNotFound}
				}
			case op == "hitLink":
				sid := strconv.Itoa(request["id"].(int))
				resCh := request["rc"].(chan response)
				if link, ok := mlm.mapLinks[sid]; ok {
					link.Hits++
					resCh <- response{value: link}
				} else {
					resCh <- response{err: ErrNotFound}
				}
			case op == "setLinkDeleted":
				sid := strconv.Itoa(request["id"].(int))
				resCh := request["rc"].(chan response)
				if link, ok := mlm.mapLinks[sid]; ok {
					dt := time.Now().UTC()
					link.DeletedAt = &dt
					resCh <- response{value: link}
				} else {
					resCh <- response{err: ErrNotFound}
				}
			case op == "removeLink":
				sid := strconv.Itoa(request["id"].(int))
				resCh := request["rc"].(chan response)
				if link, ok := mlm.mapLinks[sid]; ok {
					delete(mlm.mapIndexUrls, link.TargetUrl)
					delete(mlm.mapLinks, sid)
					resCh <- response{}
				} else {
					resCh <- response{err: ErrNotFound}
				}
			default:
				// don't panic
				continue
			}
		}
	}()
}

func (mlm *mapLinkManager) addLink(url string, expiredAt *time.Time) (int, bool, error) {
	mlm.wg.Add(1)
	defer mlm.wg.Done()
	if mlm.stop == nil {
		return -1, false, ErrClosed
	}
	request := make(request)
	request["op"] = "addLink"
	request["url"] = url
	request["expiredAt"] = expiredAt
	resCh := make(chan response)
	defer close(resCh)
	request["rc"] = resCh
	mlm.chOps <- request
	res := <-resCh
	if res.err != nil {
		return -1, false, res.err
	}
	rest, ok := res.value.(*addedResult)
	if !ok {
		return -1, false, ErrInvalidValue
	}
	return rest.id, rest.added, nil
}

func (mlm *mapLinkManager) getLink(id int) (*Link, error) {
	mlm.wg.Add(1)
	defer mlm.wg.Done()
	if mlm.stop == nil {
		return nil, ErrClosed
	}
	request := make(request)
	request["op"] = "getLink"
	request["id"] = id
	resCh := make(chan response)
	defer close(resCh)
	request["rc"] = resCh
	mlm.chOps <- request
	res := <-resCh
	if res.err != nil {
		return nil, res.err
	}
	return res.value.(*Link), res.err
}

//func (mlm *mapLinkManager) getLinks() ([]Link, error) {
//	mlm.wg.Add(1)
//	defer mlm.wg.Done()
//	if mlm.stop == nil {
//		return nil, ErrClosed
//	}
//	request := make(map[string]interface{})
//	request["op"] = "getLinks"
//	resCh := make(chan response)
//	defer close(resCh)
//	request["rc"] = resCh
//	mlm.chOps <- request
//	res := <-resCh
//	return res.value.([]Link), res.err
//}

func (mlm *mapLinkManager) hitLink(id int) (*Link, error) {
	mlm.wg.Add(1)
	defer mlm.wg.Done()
	if mlm.stop == nil {
		return nil, ErrClosed
	}
	request := make(request)
	request["op"] = "hitLink"
	request["id"] = id
	resCh := make(chan response)
	defer close(resCh)
	request["rc"] = resCh
	mlm.chOps <- request
	res := <-resCh
	if res.err != nil {
		return nil, res.err
	}
	return res.value.(*Link), res.err
}

func (mlm *mapLinkManager) setLinkDeleted(id int) error {
	mlm.wg.Add(1)
	defer mlm.wg.Done()
	if mlm.stop == nil {
		return ErrClosed
	}
	request := make(request)
	request["op"] = "setLinkDeleted"
	request["id"] = id
	resCh := make(chan response)
	defer close(resCh)
	request["rc"] = resCh
	mlm.chOps <- request
	return (<-resCh).err
}

func (mlm *mapLinkManager) removeLink(id int) error {
	mlm.wg.Add(1)
	defer mlm.wg.Done()
	if mlm.stop == nil {
		return ErrClosed
	}
	request := make(request)
	request["op"] = "removeLink"
	request["id"] = id
	resCh := make(chan response)
	defer close(resCh)
	request["rc"] = resCh
	mlm.chOps <- request
	return (<-resCh).err
}

func (mlm *mapLinkManager) Done() <-chan struct{} {
	return mlm.completed
}

func (mlm *mapLinkManager) Err() error {
	return mlm.err
}

func (mlm *mapLinkManager) save() {
	mlm.wg.Wait()
	if mlm.path != "" {
		if file, err := os.OpenFile(mlm.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
			mlm.err = err
		} else {
			if err := json.NewEncoder(file).Encode(mlm.mapLinks); err != nil {
				mlm.err = err
			}
		}
	}
}
