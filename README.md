# dsdb

お好きなデータ構造をTCPサーバーでラップします

```go
import (
	"net"

	"github.com/Mojashi/dsdb/database"
	"github.com/Mojashi/dsdb/datastructures"
)
func main() {

	listener, err := net.Listen("tcp", "localhost:5003")
	if err != nil {
		panic(err)
	}

	db := database.MakeDB()
	db.Register(datastructures.Trie{})
	db.Register(datastructures.SegmentTree{})
	db.Run(listener)
}

```

とりあえずnc localhost 5003でアクセス

- :make [データ構造名] [テーブル名] [引数...] //テーブルを作成
- :save //データベースをダンプ
- :load //データベースをロード
- [テーブル名] [関数名] [引数...] //クエリ

```
:make SegmentTree seg 10
seg Update 1 2
ret:success
seg Update 3 5
ret:success
seg Update 13 1
err:index out of range
:save
ret:success
```

## 与えるデータ構造について
- Init関数が登録してあればテーブル作成時にそれを実行します
- メンバ関数はポインタに対するメンバ関数を含めて全部見て、引数がintかstringのみならクエリの種類として登録します
- 返り値はerrorのみ、または(int, error), (string, error), (bool, error)のどれかじゃないとダメです
- save,loadはgobを使っているのでGobEncode,GobDecodeでダンプの仕方を決められます
  - なので、基本的にフィールドは全部publicで
  

## TODO
- RLockを使うように設定できるようにする
- saveのディレクトリとかポートとかを設定する手段を用意
