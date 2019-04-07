# DBUTIL

SQLDBを利用する際に考慮すべき共通の要素をまとめたもの

- 接続時に注意すべきコンフィグ設定と入力のためのyamlテンプレート
  - Connection Pool
- Timezone情報のない時間型カラムの取り扱い
- gorp/sequerri ライブラリのプレースホルダー設定


```go
import (
  _ "github.com/denisenkom/go-mssqldb"
  _ "gopkg.in/goracle.v2"
  _ "github.com/go-sql-driver/mysql"
)
```


### Connection Poolに関して

GoのSQLはコネクションプールが組み込まれている。
そのまま使うこともできるが、正しく設定をしなければパフォーマンスが出ない。
詳細は以下の記事を参照してください。

- @alexedwards [Configuring sql.DB for Better Performance](https://www.alexedwards.net/blog/configuring-sqldb)
- @methane [Re: Configuring sql.DB for Better Performance](http://dsas.blog.klab.org/archives/2018-02/configure-sql-db.html)


Recommended settings

> 推奨する sql.DB の設定
> - SetMaxOpenConns() は必ず設定する。負荷が高くなってDBの応答が遅くなったとき、新規接続してさらにクエリを投げないようにするため。できれば負荷試験をして最大のスループットを発揮する最低限のコネクション数を設定するのが良いが、負荷試験をできない場合も max_connection やコア数からある程度妥当な値を判断するべき。
> - SetMaxIdleConns() は SetMaxOpenConns() 以上に設定する。アイドルな接続の解放は SetConnMaxLifetime に任せる。
> - SetConnMaxLifetime() は最大接続数 × 1秒 程度に設定する。多くの環境で1秒に1回接続する程度の負荷は問題にならない。1時間以上に設定したい場合はインフラ／ネットワークエンジニアによく相談すること。
> 
> @methane [Re: Configuring sql.DB for Better Performance](http://dsas.blog.klab.org/archives/2018-02/configure-sql-db.html)(閲覧日:2019-04-07)



```yaml
name: mysql
driver: mysql
host: localhost:3306
user: user
pass: pass
database: reference
tz: Asia/Tokyo
max_conn: 4
max_idle_conn: 4
max_lifetime: 10
```

### DSN Parameter

#### MySQL

- user
- pass
- host: `host:port` e.g. `tcp(localhost:3306)`
- scheme

name|description
:--|:--
charset|set the charset used for client-server interaction
collation|set the collation used for client-server interaction
clientFoundRows|true/false. UPDATE to return the number of matching rows instead
columnsWithAlias|true/false. return column name withAlias Header `<table alias>.<column name>`
interpolateParams|true/false. use prepared statements
parseTime|true/false. 時間のrowをtime.Time型に変換するかどうか
loc|time.Time型のデフォルトのLocationを指定する. sessionのtimezoneは変わらないのでSystemVariablesの指定が必要
maxAllowedPacket|最大送信パケットサイズ default: 4MiB
readTimeout|I/O read timeout.
timeout|dial timeout.
writeTimeout|writeTimeout.
rejectReadOnly|true/false. failover時にread_onlyにつながったコネクションを削除するかどうか
System Variables|`<string_var>=%27<value>%27`, `<enum_var>=<value>`, `<boolean_var>=<value>` e.g. `tx_isolation=%27REPEATABLE-READ%27`, `time_zone=%27Europe%2FParis%27`

#### MSSQLDB

- user
- pass
- host: `host/instance` e.g. `localhost/sqlserver`
- host: `host:port` e.g. `localhost:1433`

name|description
:--|:--
database|Default Connect Database
connection timeout|in seconds
dial timeout|in seconds
encrypt|[disable|false|true] default:false = ログインパケット以外は暗号化しない
app name|default:go-mssqldb
keepAlive|in seconds; 0 to disable (default is 30)
failoverpartner|host or host\instance
failoverport|default 1433
packet size|in bytes; 512 to 32767 (default is 4096)
log|logging flags; 0->63. 1: errors
TrustServerCertificate|false/true
certificate|false/true
hostNameInCertificate|
ServerSPN|
Workstation ID|
ApplicationIntent|

#### GORACLE

- user
- pass
- host: `[//]host[:port][/service_name][:server][/instance_name]`, e.g. `localhost/oracle`

name|description
:--|:--
sysdba|Privileged connection
sysoper|Privileged connection
poolMinSessions|the minimum number of connections in the Oracle Access Manager(OAM) Server connection pool.
poolMaxSessions|the maximum number of connections in the OAM Server connection pool.
poolIncrement|プールの増加率?
connectionClass=POOLED|
standaloneConnection=0|Performance test?
enableEvents=0|操作イベントの取得?
heterogeneousPool=0|権限の違うセッションをプールするかどうか。セッション再確保にも認証情報を使うようになる?
prelim=0|Oracleハングアップ時に使う。ログイン時のリソース確保などがバイパスされるので、接続できる可能性が高くなる