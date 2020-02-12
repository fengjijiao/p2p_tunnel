package main

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
    // os.File.Read(), io.ReadFull() 和
    // io.ReadAtLeast() 在读取之前都需要一个固定大小的byte slice。
    // 但ioutil.ReadAll()会读取reader(这个例子中是file)的每一个字节，然后把字节slice返回。
    data, err := ioutil.ReadAll(file)
    if err != nil {
        log.Println(err)
        return nil
    }
    log.Printf("Data as hex: %x\n", data)
    log.Printf("Data as string: %s\n", data)
    log.Printf("Number of bytes read:%d\n", len(data))
    return data
}

func existsFile(filepath string) bool {
	// 文件不存在则返回error
    fileInfo, err := os.Stat(filepath)
    if err != nil {
        if os.IsNotExist(err) {
            log.Println("File does not exist.")
            return false
        }
    }
    log.Println("File does exist. File information:")
    log.Println(fileInfo)
    return true
}

func getBaseDirPath() string {
	str, _ := os.Getwd()
	return str
}