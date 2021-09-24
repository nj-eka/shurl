package errs

type Kind uint32

// todo: make kind values as flags so that masks can be used to classify compound errors
// const KindOther Kind = 1 << (32 - 1 - iota)
const (
	KindOther        Kind = iota // Unclassified error. This value is not printed in the error message.
	KindTransient                // Transient error  todo: use prev Error values
	KindInterrupted              // Interrupted ( some kind of inconsistency )
	KindInvalidValue             // Invalid value for this type of item.
	KindIO                       // External I/O error such as network failure.
	KindServer                   // Http server error
	KindRouter                   // Router error
	KindStore                    // Any kind of store failures
	KindTokenizer                // Any kind of tokenizer failures
	KindInternal                 // Internal error (for current errs pipeline impl this kind should be last in this list so that len(Kinds) = int(errs.KindInternal))
)

func (k Kind) String() string {
	switch k {
	case KindOther:
		return "other"
	case KindInvalidValue:
		return "invalid value"
	case KindInterrupted:
		return "interrupted"
	case KindIO:
		return "I/O"
	case KindServer:
		return "http server"
	case KindRouter:
		return "router"
	case KindStore:
		return "store"
	case KindTokenizer:
		return "tokenizer"
	case KindInternal:
		return "internal"
	case KindTransient:
		return "transient"
	}
	return "unknown"
}
