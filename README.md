# testProject_Golang

![img.png](images/img.png)
![img.png](images/img1.png)
Я сделал по другому задачу, так как возникли возможные изменения в API, на которые нужно было использовать. В примере показано что в tariff_description полное описание доставки, а тут либо пустой либо имеется краткое описание

В ходе выполнения задачи у меня возникла проблема с получением токена. Я не мог получить токен при использовании "application/json" в качестве Content-Type.
![img.png](images/img2.png)
![img.png](images/img3.png)
Однако, когда я попробовал использовать "application/x-www-form-urlencoded", я смог получить токен.

### Сортировка с флагами test
``` go
go run main.go AccountInfo.go --tariff-name "дверь-дверь" --test true
```

``` go
go run main.go AccountInfo.go --tariff-name "Экспресс" --test true
```

``` go
go run main.go AccountInfo.go --tariff-description "Экспресс-доставка" --tariff-description "склад-склад" --test true
```

Данные сайта Account, Secure password, API URL лежит на заигнорированном файле AccountInfo.go const(User, Password, ApiURL)