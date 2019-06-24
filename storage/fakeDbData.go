package storage

import (
	"encoding/json"
	"fmt"
	C "gamelink-go/common"
	"github.com/dustinkirkland/golang-petname"
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	letters         = "0123456789"
	goroutinesCount = 8  //8 горутин. По 1 горутине на поток проца.
	friendsCount    = 10 //Минимальное кол-во гарантированных друзей
)

//AddFakeUser - userdata goroutines start loop
func (dbs DBS) AddFakeUser(count int) {
	var wg sync.WaitGroup
	wg.Add(goroutinesCount)
	for j := 1; j <= goroutinesCount; j++ {
		go dbs.AddFakeUserGoroutine(j, count, &wg)
	}
	wg.Wait()
	return
}

//AddFakeUserGoroutine - prepare data and all db function
func (dbs DBS) AddFakeUserGoroutine(goroutineNumber, count int, wg *sync.WaitGroup) {
	fmt.Println(fmt.Sprintf("data addition gourutine № %d started", goroutineNumber))
	t1 := time.Now()
	firstID := count/goroutinesCount*(goroutineNumber-1) + 1
	lastID := count / goroutinesCount * goroutineNumber
	checkpoint := (lastID - firstID) / 10
	checkpointCount := 1
	if goroutineNumber == goroutinesCount {
		lastID = count
	}
	for i := firstID; i < lastID; i++ {
		if i%1000 == 0 && i != 0 {
			t1 = time.Now()
		}
		fakeSave := C.J{"id": i, "name": "iPhone9,3", "state": C.J{"Gold": 438, "SafeData": C.J{"MaxSum": 30, "LastTimeGet": "07/06/2018 06:28:40", "AmountInHour": 50}, "leavedId": -1, "LastTutor": 7, "LevelData": C.J{"Level": 3, "Progress": 30}, "TotalMoney": "441635587284", "AutoMultiplier": 1, "OpenedBoosters": 9, "LastBookingPerk": C.J{"ID": 6, "Level": 3}, "MissionsProgress": C.J{"CurrentMission": 0, "isFirstGoalDone": false, "isThirdGoalDone": false, "isSecondGoalDone": false}}}
		fakeData := C.J{
			"name":     petname.Generate(2, " "),
			"nickname": petname.Generate(1, ""),
			"sex":      "F",
			"lb1":      RandStringScore(100),
			"country":  "USA",
			"fb_id":    999999999999 - i,
			"save":     fakeSave,
			"email":    "gamelink@test.gamelink",
		}
		dataByte, err := json.Marshal(fakeData)
		if err != nil {
			fmt.Println(err)
			return
		}
		saveByte, err := json.Marshal(fakeSave)
		if err != nil {
			fmt.Println(err)
			return
		}

		lastID, err := dbs.AddFakeUserToDb(dataByte)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = dbs.AddFakeSaveToDb(lastID, saveByte)
		if err != nil {
			fmt.Println(err)
			return
		}
		if i-firstID == checkpoint*checkpointCount {
			fmt.Println(fmt.Sprintf("addition gourutine № %d checkpoint № %d......time elapsed till last checkpoint: %v", goroutineNumber, checkpointCount, time.Since(t1)))
			t1 = time.Now()
			checkpointCount++
		}
	}
	fmt.Println(fmt.Sprintf("addition gourutine № %d ENDED!!! Elapsed time: %v", goroutineNumber, t1))
	wg.Done()
	return
}

//AddFakeUserToDb - add userdata to db
func (dbs DBS) AddFakeUserToDb(dataByte []byte) (int64, error) {
	res, err := dbs.mySQL.Exec("INSERT INTO gamelink.users (data) VALUES (?)", dataByte)
	if err != nil {
		fmt.Println("error in AddFakeUser users")
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

//AddFakeSaveToDb - add savedata to db
func (dbs DBS) AddFakeSaveToDb(lastID int64, saveByte []byte) error {
	_, err := dbs.mySQL.Exec("INSERT INTO gamelink.saves (data, user_id) VALUES (?,?)", saveByte, lastID)
	if err != nil {
		fmt.Println("error in AddFakeUser saves")
		return err
	}
	return nil
}

//AddFakeFriends - friends goroutines start loop
func (dbs DBS) AddFakeFriends(count int) {
	var wgf sync.WaitGroup
	wgf.Add(goroutinesCount)
	for j := 1; j <= goroutinesCount; j++ {
		go dbs.AddFakeFriendsToDb(j, count, &wgf)
	}
	wgf.Wait()
	return
}

//AddFakeFriendsToDb - add friends to db
func (dbs DBS) AddFakeFriendsToDb(goroutineNumber, count int, wgf *sync.WaitGroup) {
	firstID := count/goroutinesCount*(goroutineNumber-1) + 1
	lastID := count / goroutinesCount * goroutineNumber
	if goroutineNumber == goroutinesCount {
		lastID = count
	}
	t0 := time.Now()
	t1 := time.Now()
	fmt.Println(fmt.Sprintf("friends addition gourutine № %d started", goroutineNumber))
	checkpoint := (lastID - firstID) / 10
	checkpointCount := 1
	for id := firstID; id <= lastID; id++ {
		for k := 0; k < friendsCount; k++ {
			randomFriendID := rand.Intn(count-1) + 1
			if randomFriendID == id {
				continue
			}
			_, err := dbs.mySQL.Exec("INSERT IGNORE INTO gamelink.friends (user_id1, user_id2) VALUES (?,?), (?,?)", id, randomFriendID, randomFriendID, id)
			if err != nil {
				fmt.Println("error in AddFakeFriends")
				fmt.Println(err)
				return
			}
		}
		if id-firstID == checkpoint*checkpointCount {
			fmt.Println(fmt.Sprintf("friends gourutine № %d checkpoint № %d......time elapsed till last checkpoint: %v", goroutineNumber, checkpointCount, time.Since(t1)))
			t1 = time.Now()
			checkpointCount++
		}
	}
	fmt.Println(fmt.Sprintf("friends gourutine № %d ENDED!!! Elapsed time: %v", goroutineNumber, time.Since(t0)))
	wgf.Done()
	return
}

//RandStringScore - function to generate random score string
func RandStringScore(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
