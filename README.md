# HTTPMUX


```bash
 curl -v -X POST -H "Content-Type: application/json" 
 --data '{"urls":["http://google.com","www.yandex.com"]}' http://127.0.0.1:5002/api

 ->

 {"urls":[["http://www.yandex.com","200 OK"],["http://google.com","200 OK"]]}
```
