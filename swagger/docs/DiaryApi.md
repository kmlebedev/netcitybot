# {{classname}}

All URIs are relative to *https://virtserver.swaggerhub.com/LEBEDEVKM/NetSchool/4.30.43656*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DiaryAssignnDetails**](DiaryApi.md#DiaryAssignnDetails) | **Get** /student/diary/assigns/{assignId} | 

# **DiaryAssignnDetails**
> DiaryAssignDetails DiaryAssignnDetails(ctx, assignId, studentId)


returns assign information

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **assignId** | **int32**|  | 
  **studentId** | **int32**|  | 

### Return type

[**DiaryAssignDetails**](diaryAssignDetails.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

