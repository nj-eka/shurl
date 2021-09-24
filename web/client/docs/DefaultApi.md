# UrlShortenerApi.DefaultApi

All URIs are relative to *http://localhost:8443*

Method | HTTP request | Description
------------- | ------------- | -------------
[**createShortUrl**](DefaultApi.md#createShortUrl) | **POST** / | Request short url (token) for target url with expiration interval (in days) setting
[**getShortUrlInfo**](DefaultApi.md#getShortUrlInfo) | **GET** /{token}/info | Get short url info
[**hitShortUrl**](DefaultApi.md#hitShortUrl) | **GET** /{token} | Redirect to target url by token



## createShortUrl

> ResponseShortUrl createShortUrl(requestShortUrl)

Request short url (token) for target url with expiration interval (in days) setting

### Example

```javascript
import UrlShortenerApi from 'url_shortener_api';

let apiInstance = new UrlShortenerApi.DefaultApi();
let requestShortUrl = new UrlShortenerApi.RequestShortUrl(); // RequestShortUrl | 
apiInstance.createShortUrl(requestShortUrl, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **requestShortUrl** | [**RequestShortUrl**](RequestShortUrl.md)|  | 

### Return type

[**ResponseShortUrl**](ResponseShortUrl.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## getShortUrlInfo

> Link getShortUrlInfo(token)

Get short url info

### Example

```javascript
import UrlShortenerApi from 'url_shortener_api';

let apiInstance = new UrlShortenerApi.DefaultApi();
let token = "token_example"; // String | short url info
apiInstance.getShortUrlInfo(token, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **token** | **String**| short url info | 

### Return type

[**Link**](Link.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## hitShortUrl

> hitShortUrl(token)

Redirect to target url by token

### Example

```javascript
import UrlShortenerApi from 'url_shortener_api';

let apiInstance = new UrlShortenerApi.DefaultApi();
let token = "token_example"; // String | hit short url (redirect to target url + increment hits)
apiInstance.hitShortUrl(token, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully.');
  }
});
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **token** | **String**| hit short url (redirect to target url + increment hits) | 

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

