package triedb

import (
	"os"
	"testing"

	crand "crypto/rand"
	mrand "math/rand"
)

type withFatal interface {
	Fatal(args ...any)
	Fatalf(string, ...any)
}

func prepareSampleDatabase(t withFatal, tdb *TrieDB) [][]byte {
	existingKeys := make(map[string]struct{}, 0)

	// about 218 seconds
	for i := 0; i < 10000*100*1; i++ {
		percent := mrand.Intn(100)
		if percent < 55 { // 55% Update
			tokenKey := make([]byte, 8)
			tokenVal := make([]byte, 8)
			{
				n, err := crand.Read(tokenKey)
				if nil != err {
					t.Fatal("unexpected random token error: ", err)
				}
				if n < 8 {
					t.Fatal("unexpected random token length: ", n)
				}
			}
			{
				n, err := crand.Read(tokenVal)
				if nil != err {
					t.Fatal("unexpected random token error: ", err)
				}
				if n < 8 {
					t.Fatal("unexpected random token length: ", n)
				}
			}
			// fmt.Println(n, tokenKey, tokenVal)
			err := tdb.Put(Path(tokenKey), tokenVal, true)
			if nil != err {
				t.Fatal("unexpected Update error: ", err)
			}
			existingKeys[bytes2String(tokenKey)] = struct{}{}
		} else if percent < 90 { // 35% Get
			for k := range existingKeys {
				v, err := tdb.Get(Path(string2Bytes(k)))
				if nil != err {
					t.Fatal("unexpected Get error: ", err)
				}
				if nil == v {
					t.Fatalf("value for key [%#v] must NOT be nil, but: %#v", k, v)
				}
				break
			}
		} else { // 10% Delete
			var deletingKey []byte
			if len(existingKeys) > 0 {
				index := mrand.Intn(len(existingKeys))
				for k := range existingKeys {
					if index <= 0 {
						deletingKey = string2Bytes(k)
						break
					}
					index--
				}
			}

			if nil != deletingKey {
				_, err := tdb.Del(Path(deletingKey))
				if nil != err {
					t.Fatal("unexpected Delete error: ", err)
				}
				delete(existingKeys, bytes2String(deletingKey))
			}
		}
	}

	bytesForExistingKeys := make([][]byte, 0)
	for k, _ := range existingKeys {
		bytesForExistingKeys = append(bytesForExistingKeys, string2Bytes(k))
	}
	return bytesForExistingKeys
}

func TestMorePressure(t *testing.T) {
	const db_path = "./triedb_pressure_test.db"
	os.RemoveAll(db_path)

	tdb, err := testPrepareTrieDB(db_path, nil)

	if nil != err {
		t.Fatal(err)
	}

	prepareSampleDatabase(t, tdb)

	tdb.testCommit()
	// tdb.GenDotFile("./test_withmorepressure.dot", false)
	testCloseTrieDB(tdb)
}

// give more pressure to do more operations
func BenchmarkMoreOperations(b *testing.B) {
	const db_path = "./triedb_benchmark_test.db"
	defer os.RemoveAll(db_path)

	tdb, err := testPrepareTrieDB(db_path, nil)

	if nil != err {
		b.Fatal(err)
	}
	existingKeys := prepareSampleDatabase(b, tdb)

	const COUNT = 10000 * 500

	randomUpdateKeys := make([][]byte, 0, COUNT)
	randomUpdateVals := make([][]byte, 0, COUNT)

	for i := 0; i < COUNT; i++ {
		tokenKey := make([]byte, mrand.Intn(16)+1)
		tokenVal := make([]byte, mrand.Intn(16)+1)
		crand.Read(tokenKey)
		crand.Read(tokenVal)
		randomUpdateKeys = append(randomUpdateKeys, tokenKey)
		randomUpdateVals = append(randomUpdateVals, tokenVal)
	}

	randomGetKeys := make([][]byte, 0, COUNT)
	for i := 0; i < COUNT; i++ {
		j := mrand.Intn(len(existingKeys))
		randomGetKeys = append(randomGetKeys, existingKeys[j])
	}

	randomDeleteKeys := make([][]byte, 0, COUNT)
	for i := 0; i < COUNT; i++ {
		j := mrand.Intn(len(existingKeys))
		randomDeleteKeys = append(randomDeleteKeys, existingKeys[j])
	}

	b.ResetTimer()
	b.Run("Update", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tokenKey := randomUpdateKeys[i]
			tokenVal := randomUpdateVals[i]
			err := tdb.Put(Path(tokenKey), tokenVal, true)
			if nil != err {
				b.Fatal(err)
			}
		}
	})
	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if len(existingKeys) > 0 {
				tokenKey := randomGetKeys[i]
				val, err := tdb.Get(Path(tokenKey))
				if nil != err {
					b.Fatal(err)
				}
				if nil == val {
					b.Fatal("unexpected nil value for key: ", tokenKey)
				}
			}
		}
	})
	b.Run("Delete", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if len(existingKeys) > 0 {
				tokenKey := randomDeleteKeys[i]
				_, err := tdb.Del(Path(tokenKey))
				if nil != err {
					b.Fatal(err)
				}
			} else {
				break
			}
		}
	})

	tdb.testCommit()
	// tdb.GenDotFile("./test_domoreoperations.dot", false)
	testCloseTrieDB(tdb)
}
