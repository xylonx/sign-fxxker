general sign in
https://mooc1.chaoxing.com/visit/stucoursemiddle?courseid=214856782&clazzid=32696035&cpi=114287792&ismooc2=1

https://mooc1.chaoxing.com/visit/stucoursemiddle?courseid=${CourseID}&clazzid=${ClassID}&cpi=${PersonID}&ismooc2=1

> var personId = $('.ClassDetail_teach .choseItem.active').attr('data-personid');
> 




GET https://mobilelearn.chaoxing.com/v2/apis/sign/getAttendInfo?activeId=4000014316792

refer https://mobilelearn.chaoxing.com/page/sign/signIn?courseId=223336235&classId=52378732&activeId=4000014340734&fid=0

```json
{
    "result": 1,
    "msg": "success",
    "data": {
        "id": 4000108383925,
        "uid": 122949003,
        "activeId": 4000014316792,
        "status": 1,
        "createtime": 1646096800000,
        "updatetime": 1646096799000,
        "name": "向宇鑫",
        "username": "",
        "isdelete": 0,
        "qasort": null,
        "sasort": 1646096799870,
        "longitude": null,
        "latitude": null,
        "distance": null,
        "distanceStr": null,
        "xxuid": 0,
        "clientip": null,
        "issend": null,
        "useragent": null,
        "type": 0,
        "confidence": null,
        "islook": 1,
        "score": null,
        "skscore": null,
        "result": null,
        "title": null,
        "iphoneContent": null,
        "ismark": 0,
        "submittime": "2022-03-01 09:06:39",
        "groupId": 0,
        "taskavg": 0.0,
        "content": null,
        "answerScore": null,
        "answerTime": null,
        "teacherGiveScore": 0,
        "isPrompted": 0,
        "mutualEvaluationId": null,
        "pingyu": null,
        "flowerCount": null,
        "schoolname": null,
        "fid": 0,
        "screenContent": null,
        "screenimages": null,
        "iframeCount": null,
        "showScore": 0,
        "teaScore": null,
        "ifGiveScore": null,
        "taskScoreRecored": null,
        "taskTeacherEvaluation": null,
        "updatetimeStr": "2022-03-01 09:06:39",
        "stuGetAnswerTime": "",
        "teaGetAnswerTime": ""
    },
    "errorMsg": null
}
```


GET https://mobilelearn.chaoxing.com/page/sign/signIn?courseId=223336235&classId=52378732&activeId=4000014340734&fid=0