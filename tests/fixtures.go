package tests

const jsonObjectTest = `{"name": "Jason", "age": 15, "address": null}`
const jsonArrayTest = `[{"name": "Jason"}, {"name":  "Chris"}]`
const jsonArrayWithNullTest = `[{"name": "Jason"}, {"name": "Chris"}, null]`
const jsonTimeTest = `{"time1": "2024-10-06T17:59:44Z", "time2": "2024-10-06T17:59:44+00:00", "time3": "Sunday, 06-Oct-24 17:59:44 UTC"}`
const jsonTimeTestInvalid = `{"time1": null, "time2": 0, "time3": "INVALID", "time4": false}`
