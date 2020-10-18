# Golang × React Chat Demo
![Demo](https://user-images.githubusercontent.com/61624833/96363076-f507c780-116c-11eb-84fa-359d45bf14da.gif)
GolangでWebsockets, Redisを使用した簡易チャットアプリ<br>
Webサーバーをscallableにするため、RedisのPubsubを活用。<br>

## Quickstart
以下をクローンしたディレクトリ内で実行
```sh
$ make
$ cd frontend
$ yarn start
```

## Usage
Makefileの内容<br>
#### `make up`
Golangのコンテナを2台作成。
#### `make down`
コンテナを停止・削除
#### `make re`
コンテナを削除してから、作り直す。
#### `make redis`
Redisを実行しているコンテナの中に入る
#### `make go1`
1つめのGolangコンテナの中に入る
#### `make go2`
2つめのGolangコンテナの中に入る

## About this project
基本的な考え方は[AWSのサンプル](https://aws.amazon.com/jp/blogs/news/how-to-build-a-chat-application-with-amazon-elasticache-for-redis/)を参考にした<br>
このプロジェクトの大まかな仕様は以下のような感じ<br>
`hub.go`<br>
Webサーバー内のユーザー・メッセージの処理を行う。Redisとの連携もここで行う。<br>
`client.go`<br>
各ユーザーが実行する処理。<br>
`redis.go`<br>
Redisに対して実行する処理。<br>
<br>
Hubはそのサーバーにアクセスしているユーザーを管理している。Hubを通して、ユーザーに情報（メッセージ・ユーザーの入退室）を送信する。<br>
Redisを使わない場合、そのサーバー内でしかWebsocketの通信は行われないため、アクセスしているサーバーが違う場合、同じチャンネルを見ていても相互の通信ができない。<br>
そこでRedisのPubsubを使い、Hub（サーバー）自身が参照したいチャンネルをSubscribeする。<br>
すると以下のような流れになる。<br>
- メッセージをRedisにPublishする
- チャンネルをSubscribeしたHub（サーバー）に対して、メッセージが渡される
- Hub内のユーザーで、同じチャンネルをもったユーザーにメッセージが振り分けられる
