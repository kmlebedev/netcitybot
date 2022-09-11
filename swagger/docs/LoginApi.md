# {{classname}}

All URIs are relative to *https://virtserver.swaggerhub.com/LEBEDEVKM/NetSchool/4.30.43656*

Method | HTTP request | Description
------------- | ------------- | -------------
[**Logindata**](LoginApi.md#Logindata) | **Get** /logindata | 
[**Prepareemloginform**](LoginApi.md#Prepareemloginform) | **Get** /prepareemloginform | 
[**Prepareloginform**](LoginApi.md#Prepareloginform) | **Get** /prepareloginform | 

# **Logindata**
> LoginData Logindata(ctx, )


returns all login data

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**LoginData**](LoginData.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **Prepareemloginform**
> PrepareEmLoginForm Prepareemloginform(ctx, optional)


returns all prepareemloginform

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***LoginApiPrepareemloginformOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a LoginApiPrepareemloginformOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **cacheVer** | **optional.String**|  | 

### Return type

[**PrepareEmLoginForm**](PrepareEmLoginForm.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **Prepareloginform**
> PrepareLoginForm Prepareloginform(ctx, optional)


returns all prepareloginform

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***LoginApiPrepareloginformOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a LoginApiPrepareloginformOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **cacheVer** | **optional.String**|  | 

### Return type

[**PrepareLoginForm**](PrepareLoginForm.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

