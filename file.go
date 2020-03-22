package p2p

import (
	"os"
	"log"
	"io/ioutil"
)

func test_file() {
	createFile("test.txt", []byte("test content"))
	deleteFile("test.txt")
	appendFile("test.txt", []byte("test content"))
}

func deleteFile(filepath string) bool {
	err := os.Remove(filepath)
    if err != nil {
        log.Println(err)
        return false
    }else {
    	return true
    }
}

func createFile(filepath string, byteSlice []byte) bool {
	// 可写方式打开文件
    file, err := os.OpenFile(
        filepath,
        os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
        0666,
    )
    if err != nil {
        log.Println(err)
        return false
    }
    defer file.Close()
    // 写字节到文件中
    bytesWritten, err := file.Write(byteSlice)
    if err != nil {
        log.Println(err)
        return false
    }
    log.Printf("Wrote %d bytes.\n", bytesWritten)
    return true
}

func appendFile(filepath string, byteSlice []byte) bool {
    file, err := os.OpenFile(filepath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
    if err != nil {
        log.Printf("An error occurred with file opening or creation\n")
        return false
    }
    defer file.Close()
    file.Write(byteSlice)
    return true
}

func readFile(filepath string) []byte {
	file, err := os.Open(filepath)
    if err != nil {
        log.Println(err)
        return nil
    }
    data, err := ioutil.ReadAll(file)
    if err != nil {
        log.Println(err)
        return nil
    }
    return data
}

func existsFile(filepath string) bool {
	// 文件不存在则返回error
    _, err := os.Stat(filepath)
    if err != nil {
        if os.IsNotExist(err) {
            return false
        }
    }
    return true
}

func getBaseDirPath() string {
	str, _ := os.Getwd()
	return str
}