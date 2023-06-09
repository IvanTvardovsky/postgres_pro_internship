## Postgres Pro - Go стажёр тестовое задание 

Текст задания в task_README.md

#### Инструкция по запуску:

* Создать базу данных PostgreSQL с двумя таблицами:
1. 
```sql 
CREATE TABLE changes (
   id SERIAL PRIMARY KEY,
   file VARCHAR(255),
   date TIMESTAMP
   );
```
2. 
```sql
CREATE TABLE executed_commands (
  id SERIAL PRIMARY KEY,
  command VARCHAR(255),
  date TIMESTAMP,
  change_id INTEGER,
  FOREIGN KEY (change_id) REFERENCES changes (id)
);
```
* Задайте в файле **config.yml**:
1. В разделе **storage** все необходимые данные от базы данных.
2. В разделе **directories** все директории, наблюденте над которыми вы хотите установить (настройки include_regexp, exclude_regexp, log_file не реализованы ввиду недостатка времени). Вы можете установить наблюдение сразу за несколькими директориями, перечисляя их через "-". 
* Установив все необходимые пакеты, используемые проектом, запустите **/cmd/main/app.go**

### Возникшие вопросы

* Как отслеживать изменения файлов?

При использовании метода polling производительность может страдать, особенно если в директории находится большое количество файлов.
Функция os.File.Readdir() позволяет получать информацию об изменениях в реальном времени, но может столкнуться с проблемой блокировки файлов, если файл был открыт другим процессом.
Использование fsnotify.NewWatcher() может снизить нагрузку на процессор, поскольку он будет получать уведомления только о событиях в файловой системе, которые произошли после регистрации Watcher. Хоть у этого метода и есть свои недостатки, он является лучшим выбором в нашем случае.

* Как хранить изменения файлов и запусков?

Был выбран довольно простой, но экономный вариант - для изменений файлов хранятся только путь к файлу и дата изменения, а так же присваивается уникальный идентификатор. Для списка запусков имеется дополнительно поле, которые отсылает к id изменения, из-за которого была вызвана команда.