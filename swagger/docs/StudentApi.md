# {{classname}}

All URIs are relative to *https://virtserver.swaggerhub.com/LEBEDEVKM/NetSchool/1.0.0*

Method | HTTP request | Description
------------- | ------------- | -------------
[**StudentDiary**](StudentApi.md#StudentDiary) | **Get** /student/diary | 

# **StudentDiary**
> Diary StudentDiary(ctx, studentId, optional)


returns all assignments

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **studentId** | **string**|  | 
 **optional** | ***StudentApiStudentDiaryOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a StudentApiStudentDiaryOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **weekStart** | **optional.String**|  | 
 **weekEnd** | **optional.String**|  | 
 **withLaAssigns** | **optional.Bool**|  | 
 **withPastMandatory** | **optional.Bool**|  | 
 **yearId** | **optional.Int32**|  | 

### Return type

[**Diary**](Diary.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

