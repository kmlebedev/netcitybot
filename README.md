# netcitybot телеграм бот для "Сетевой Город. Образование"

Базовый функционал по пересылки домашки в групповой чат класс.
Для этого нужно добавить, создать своего телеграмм бота и добавить его в группу класса.
Далее скачать и запустить бинарь https://github.com/kmlebedev/netcitybot/releases через переменно окружение передав ссылку до сервера, логин, пароль и ChatID канала.
Профит в том, что приходят уведомления, когда приходит домашка и в случае с перебоями работы сервера электронного дневника всегда есть под рукой задания со сложениями.

![Screenshot 2022-09-14 at 20 41 02](https://user-images.githubusercontent.com/9497591/190201195-276ee759-4b92-4f5c-bb31-a196b18246a0.png)

# Быстрый страт
## Установка бота
1. Необходимо настроить переменное окружение
```
NETCITY_URL=http://192.168.1.1  # URL для сервера Сетевой Город. Образование
NETCITY_STUDENT_IDS=71111,72222 # Id учеников чья домашка будет пересылаться в чат класс. Обычно это мальчик и девочка чью группы не пересекаются
NETCITY_SCHOOL=МБОУ СОШ №1      # Образовательная организация 
NETCITY_USERNAME=ИвановИ        # Любой логин
NETCITY_PASSWORD=123456         # Пароль
NETCITY_YEAR_ID=                # Опционально, если клиент не в состоянии сам получить id
NETCITY_URLS=http://192.168.1.1 # Бот подготовит данные для диалого со входом в дневник
BOT_API_TOKEN=xxxxxxxxxxxxxxxxx # Как создать бота https://tlgrm.ru/docs/bots#kak-sozdat-bota
BOT_CHAT_ID=170000000           # Чат класса для пересылки домашки                                                
```
2. Скачать и запустить приложение 
```
netcitybot
INFO[0000] CURRYEAR value: 206, 2021/2022               
INFO[0000] Sync years: 11, classes: 42, students: 1503     
INFO[0000] LoopPullingOrder chatId: xxxxx, yearId: 235
2022/09/13 11:44:58 Endpoint: getUpdates, params: map[allowed_updates:null timeout:60]
```
## Запуск docker сервиса 
1. Скачать docker-compose.yaml:
```bash
curl -fsSL https://raw.githubusercontent.com/kmlebedev/netcitybot/main/docker/docker-compose.yml -o docker-compose.yml
```
2. Установить переменные:
```bash
echo "NETCITY_URL=http://192.168.1.1
NETCITY_STUDENT_IDS=71111,72222
NETCITY_SCHOOL=МБОУ СОШ №1
NETCITY_USERNAME=ИвановИ
NETCITY_PASSWORD=123456
NETCITY_YEAR_ID=
NETCITY_URLs=http://192.168.1.1,http://192.168.1.2
BOT_API_TOKEN=xxxxxxxxxxxxxxxxx
BOT_CHAT_ID=170000000" > .env_hobby
```
3. Запустить сервис:
```bash
docker-compose --env-file .env_hobby -f docker-compose.yml up -d
```

# Доккументация по публичному Web API NetSchool
*https://app.swaggerhub.com/apis/LEBEDEVKM/NetSchool/4.30.43656*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*AssignmentApi* | [**AssignmentTypes**](docs/AssignmentApi.md#assignmenttypes) | **Get** /grade/assignment/types |
*DiaryApi* | [**DiaryAssignnDetails**](docs/DiaryApi.md#diaryassignndetails) | **Get** /student/diary/assigns/{assignId} |
*LoginApi* | [**Logindata**](docs/LoginApi.md#logindata) | **Get** /logindata |
*LoginApi* | [**Prepareemloginform**](docs/LoginApi.md#prepareemloginform) | **Get** /prepareemloginform |
*LoginApi* | [**Prepareloginform**](docs/LoginApi.md#prepareloginform) | **Get** /prepareloginform |
*StudentApi* | [**StudentDiary**](docs/StudentApi.md#studentdiary) | **Get** /student/diary |
*StudentApi* | [**StudentDiaryInit**](docs/StudentApi.md#studentdiaryinit) | **Get** /student/diary/init |

# Todo Публичны бот для всех
Но ничего не мешает расширить функционал, до общего для всех бота.  Пока план такой.
Но при первом старте необходимо будет ввести:
1. Выбрать город => школу
2. Если города или школы нет, то ввести ссылку на электронный дневник с возвратом на шаг 1
3. Ввести логин и пароль
4. Далее данные синхронизируются

И станет доступно меню и команды
## Команды:
1. login - повторная авторизация при смене пароля
2. logout - выход из дневника
3. subscribe - подписаться/отписаться на выбранные события в дневнике
4. forward - включить/выключить пересылку заданий по выбранным предметам
5. track - отслеживать успеваемость, сделанные уроки, средний балл, спорные оценки, прогноз вероятности пройти конкурс в 10-й класс.

## Меню:
1. Домашние задания на сегодня или на конкретную неделю (inline кнопки дз: сделано/не сделано)
2. Отчеты по успеваемости
3. Непрочитанные сообщения и анонсы
4. Написать и отправить сообщение

Рабочие сервера Сетевой Город. Образование:
* МО
  * http://www.netschool.pavlovo-school.ru/  (v5.5.61111.35)
* Челябинск:
  * https://sgo.edu-74.ru/ (v5.9.62423.47)
* Екатеринбург: 
  * http://188.226.50.152/ (v4.30.43656.19) - Ленинский район 
  * http://schoolroo.ru/ (v5.6.61460.74) - Октябрьский район
  * https://sg.lyceum130.ru/ (v5.0.60380.1220) - Кировский район
  * http://5.165.26.113  (v3.10.31549) - Чкаловский район
  * http://school23ekb.ru/ (v5.5.61111.35)- МАОУ СОШ № 23
  * https://my.shkola79.ru/ (5.10.35574.782)- МАОУ СОШ № 79 
* СО http://xn---66-eddggda1bzcdazfq.xn--p1ai/
  * http://31.28.113.161/ (v4.75.56652.961) - Красноуфимск 
  * http://94.190.51.157/ - Первоуральск
* http://spo.cit73.ru/ (v3.5.0.97) - Самара 
* https://sgo.egov66.ru/ (v5.0.59442.919) - Нижний Тагил
* https://net-school.cap.ru/ (v5.9.62423.47)
* https://sgo1.edu71.ru/ (v5.0.59442.919)
* https://sgo.e-mordovia.ru/ (v5.9.62423.47)
* https://netschool.eduportal44.ru/ (v5.9.62423.47)
* http://94.190.51.157/ (v5.7.61806.72)
* https://dnevnik-kchr.ru/ (v4.55.51003.990) - Черкесск
* https://poo.e-yakutia.ru/ (v3.7.0.19)
* https://sgo.e-yakutia.ru/ (v5.9.62423.47) - Якутск
* https://sgo.rso23.ru/ (5.9.62423.47) - Краснодар

# Аналоги ботов
* https://github.com/ser-ogonkov/dnevnik_bot
* https://github.com/nickname123456/BotNetSchool
