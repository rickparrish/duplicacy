// Copyright (c) Acrosync LLC. All rights reserved.
// Free for personal use and commercial trial
// Commercial use requires per-user licenses available from https://duplicacy.com

package duplicacy

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"testing"

	crypto_rand "crypto/rand"
	"math/rand"
)

func TestOneDriveClient(t *testing.T) {
	setTestingT(t)
	SetLoggingLevel(DEBUG)
	
	isBusiness := false
	isShared := true
	
	testDir := "test"
	if isShared {
		testDir = "sharedtest"
	}
	
	tokenFile := "one-token.json"
	if isBusiness {
		tokenFile = "odb-token.json"
	}

	oneDriveClient, err := NewOneDriveClient(tokenFile, isBusiness)
	if err != nil {
		t.Errorf("Failed to create the OneDrive client: %v", err)
		return
	}

	oneDriveClient.TestMode = true
	
	oneDriveClient.DetectSharedStorage(testDir)

	existingFiles, err := oneDriveClient.ListEntries("")
	for _, file := range existingFiles {
		fmt.Printf("name: %s, isDir: %t\n", file.Name, len(file.Folder) != 0)
	}

	testID, _, _, err := oneDriveClient.GetFileInfo(testDir)
	if err != nil {
		t.Errorf("Failed to list the test directory: %v", err)
		return
	}
	if testID == "" {
		if isShared {
			t.Errorf("Test directory must exist for shared testing: " + testDir)
			return
		} else {
			err = oneDriveClient.CreateDirectory("", testDir)
			if err != nil {
				t.Errorf("Failed to create the test directory: %v", err)
				return
			}
		}
	}

	test1ID, _, _, err := oneDriveClient.GetFileInfo(testDir + "/test1")
	if err != nil {
		t.Errorf("Failed to list the test1 directory: %v", err)
		return
	}
	if test1ID == "" {
		err = oneDriveClient.CreateDirectory(testDir, "test1")
		if err != nil {
			t.Errorf("Failed to create the test1 directory: %v", err)
			return
		}
	}

	test2ID, _, _, err := oneDriveClient.GetFileInfo(testDir + "/test2")
	if err != nil {
		t.Errorf("Failed to list the test2 directory: %v", err)
		return
	}
	if test2ID == "" {
		err = oneDriveClient.CreateDirectory(testDir, "test2")
		if err != nil {
			t.Errorf("Failed to create the test2 directory: %v", err)
			return
		}
	}

	numberOfFiles := 20
	maxFileSize := 64 * 1024

	for i := 0; i < numberOfFiles; i++ {
		content := make([]byte, rand.Int()%maxFileSize+1)
		_, err = crypto_rand.Read(content)
		if err != nil {
			t.Errorf("Error generating random content: %v", err)
			return
		}

		hasher := sha256.New()
		hasher.Write(content)
		filename := hex.EncodeToString(hasher.Sum(nil))

		fmt.Printf("file: %s\n", filename)

		err = oneDriveClient.UploadFile(testDir + "/test1/"+filename, content, 100)
		if err != nil {
			/*if e, ok := err.(ACDError); !ok || e.Status != 409 */ {
				t.Errorf("Failed to upload the file %s: %v", filename, err)
				return
			}
		}
	}

	entries, err := oneDriveClient.ListEntries(testDir + "/test1")
	if err != nil {
		t.Errorf("Error list randomly generated files: %v", err)
		return
	}

	for _, entry := range entries {
		err = oneDriveClient.MoveFile(testDir + "/test1/"+entry.Name, testDir + "/test2")
		if err != nil {
			t.Errorf("Failed to move %s: %v", entry.Name, err)
			return
		}
	}

	entries, err = oneDriveClient.ListEntries(testDir + "/test2")
	if err != nil {
		t.Errorf("Error list randomly generated files: %v", err)
		return
	}

	for _, entry := range entries {
		readCloser, _, err := oneDriveClient.DownloadFile(testDir + "/test2/" + entry.Name)
		if err != nil {
			t.Errorf("Error downloading file %s: %v", entry.Name, err)
			return
		}

		hasher := sha256.New()
		io.Copy(hasher, readCloser)
		hash := hex.EncodeToString(hasher.Sum(nil))

		if hash != entry.Name {
			t.Errorf("File %s, hash %s", entry.Name, hash)
		}

		readCloser.Close()
	}

	for _, entry := range entries {

		err = oneDriveClient.DeleteFile(testDir + "/test2/" + entry.Name)
		if err != nil {
			t.Errorf("Failed to delete the file %s: %v", entry.Name, err)
			return
		}
	}

}
