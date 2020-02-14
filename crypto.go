package main

import (
    "fmt"
    "log"
    "github.com/golang/protobuf/proto"
    "github.com/google/tink/go/aead"
    "github.com/google/tink/go/keyset"
    "github.com/google/tink/go/insecurecleartextkeyset"
    "encoding/base64"
    "bytes"
)

var kh *keyset.Handle

func test_crypto() {
    kh, _ = keyset.NewHandle(aead.AES128GCMKeyTemplate())
    // first test
    a, err := aead.New(kh)
    if err != nil {
        log.Fatal(err)
    }

    ct, err := a.Encrypt([]byte("this data needs to be encrypted"), []byte("associated data"))
    if err != nil {
        log.Fatal(err)
    }

    pt, err := a.Decrypt(ct, []byte("associated data"))
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Cipher text: %s\nPlain text: %s\n", ct, pt)
    // first test
    exportedPriv := &keyset.MemReaderWriter{}
    if err := insecurecleartextkeyset.Write(kh, exportedPriv); err != nil {
            log.Fatalf("unexpected error writing keyset: %v", err)
    }
    ksPriv, _ := proto.Marshal(exportedPriv.Keyset)
    bksPriv := base64.StdEncoding.EncodeToString(ksPriv)
    createFile("priv.txt", []byte(bksPriv))
    fmt.Println(bksPriv)
    bksPub := readFile("priv.txt")
    ksPub, _ := base64.StdEncoding.DecodeString(string(bksPub))
    khPub, err := insecurecleartextkeyset.Read(keyset.NewBinaryReader(bytes.NewReader(ksPub)))
    // second test
    a, err = aead.New(khPub)
    if err != nil {
        log.Fatal(err)
    }

    ct, err = a.Encrypt([]byte("this data needs to be encrypted"), []byte("associated data"))
    if err != nil {
        log.Fatal(err)
    }

    pt, err = a.Decrypt(ct, []byte("associated data"))
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Cipher text: %s\nPlain text: %s\n", ct, pt)
    // second test
}
//JSON文件key
func patchKeysetHandleFromJSON(filepath string) {
    buf := new(bytes.Buffer)
    w := keyset.NewJSONWriter(buf)
    r := keyset.NewJSONReader(buf)
    if(existsFile(filepath)){
        bksPub := readFile(filepath)
        r = keyset.NewJSONReader(bytes.NewBufferString(string(bksPub)))
        ks, _ := r.Read()
        //fmt.Printf("%+v\n", ks)
        kh, _ = insecurecleartextkeyset.Read(&keyset.MemReaderWriter{Keyset: ks})//ks-->kh
    }else{
        kh, _ = keyset.NewHandle(aead.AES128GCMKeyTemplate())
        ks := insecurecleartextkeyset.KeysetMaterial(kh)//kh-->ks
        //fmt.Printf("%+v\n", ks)
        w.Write(ks)
        //fmt.Printf("%+v\n", buf)
        createFile(filepath, buf.Bytes())
    }
    //fmt.Printf("%+v\n", kh)
}
//普通文件key
func patchKeysetHandle(filepath string) {
    if(existsFile(filepath)){
        bksPub := readFile(filepath)
        ksPub, _ := base64.StdEncoding.DecodeString(string(bksPub))
        kh, _ = insecurecleartextkeyset.Read(keyset.NewBinaryReader(bytes.NewReader(ksPub)))
    }else{
        kh, _ = keyset.NewHandle(aead.AES128GCMKeyTemplate())
        exportedPriv := &keyset.MemReaderWriter{}
        if err := insecurecleartextkeyset.Write(kh, exportedPriv); err != nil {
            log.Fatalf("unexpected error writing keyset: %v", err)
        }
        ksPriv, _ := proto.Marshal(exportedPriv.Keyset)
        bksPriv := base64.StdEncoding.EncodeToString(ksPriv)
        createFile(filepath, []byte(bksPriv))
    }
    //fmt.Printf("%+v\n", kh)
}
// cryptOfAEAD
// AEAD加密&解密
func cryptOfAEAD(text []byte, key []byte, isDecrypt bool) []byte {
    //log.Printf("%+v\n", kh)
    //log.Printf("original：’%v‘\n", text)
    aen, err := aead.New(kh)
    if err != nil {
        log.Fatal(err)
    }
    if(!isDecrypt){
        ct, err := aen.Encrypt(text, key)
        if err != nil {
            log.Fatal(err)
        }
        //log.Printf("Encrypted results：’%v‘\n", ct)
        return []byte(base64.StdEncoding.EncodeToString(ct))
    }else{
        textdecoding, _ := base64.StdEncoding.DecodeString(string(text))
        text = []byte(textdecoding)
        pt, err := aen.Decrypt(text, key)
        if err != nil {
            log.Fatal(err)
        }
        //log.Printf("Decryption result：‘%v’\n", pt)
        return pt
    }
}

func cryptOfAEADString(text []byte, key []byte, isDecrypt bool) string {
    //log.Printf("%+v\n", kh)
    //log.Printf("original：’%v‘\n", text)
    aen, err := aead.New(kh)
    if err != nil {
        log.Fatal(err)
    }
    if(!isDecrypt){
        ct, err := aen.Encrypt(text, key)
        if err != nil {
            log.Fatal(err)
        }
        //log.Printf("Encrypted results：’%v‘\n", ct)
        return base64.StdEncoding.EncodeToString(ct)
    }else{
        textdecoding, _ := base64.StdEncoding.DecodeString(string(text))
        text = []byte(textdecoding)
        pt, err := aen.Decrypt(text, key)
        if err != nil {
            log.Fatal(err)
        }
        //log.Printf("Decryption result：‘%v’\n", pt)
        return string(pt)
    }
}

func checkCryptOfAEAD(text []byte, key []byte, isDecrypt bool) bool {
    aen, err := aead.New(kh)
    if err != nil {
        log.Println(err)
        return false
    }
    if(!isDecrypt){
        _, err := aen.Encrypt(text, key)
        if err != nil {
            log.Println(err)
            return false
        }
        return true
    }else{
        textdecoding, _ := base64.StdEncoding.DecodeString(string(text))
        text = []byte(textdecoding)
        _, err := aen.Decrypt(text, key)
        if err != nil {
            log.Println(err)
            return false
        }
        return true
    }
}