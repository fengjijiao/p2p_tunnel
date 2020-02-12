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
}
// cryptOfAEAD
// AEAD加密&解密
func cryptOfAEAD(text []byte, key []byte, isDecrypt bool) []byte {
    //log.Printf("%+v\n", kh)
    //log.Printf("原文：’%v‘\n", text)
    aen, err := aead.New(kh)
    if err != nil {
        log.Fatal(err)
    }
    if(!isDecrypt){
        ct, err := aen.Encrypt(text, key)
        if err != nil {
            log.Fatal(err)
        }
        //log.Printf("加密结果：’%v‘\n", ct)
        return ct
    }else{
        pt, err := aen.Decrypt(text, key)
        if err != nil {
            log.Fatal(err)
        }
        //log.Printf("解密结果：‘%v’\n", pt)
        return pt
    }
}