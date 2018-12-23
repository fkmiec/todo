package todolist

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

type TodoSync struct {
	config        *Config
	store         Store
	addedTodos    []*Todo
	modifiedTodos []*Todo
	deletedTodos  []*Todo
	Checkpoint    *Todo
	Backlog       *TodoList
	Remote        *TodoList
	Local         *TodoList
}

func NewTodoSync(cfg *Config, s Store) *TodoSync {
	return &TodoSync{config: cfg, store: s, Backlog: &TodoList{}, Remote: &TodoList{}, Local: &TodoList{}}
}

func (s *TodoSync) Sync(verbose bool) error {
	if s.config.SyncFilepath == "" {
		return errors.New("No sync.filepath defined in .todorc config file")
	}
	syncFilepath := s.config.SyncFilepath
	origSyncFilepath := ""
	encryptionPassphrase := ""
	//Check if encryption desired. If so, create tmp file for decrypted content
	//If no encryptionPassphrase value, assume no encryption to be used
	if s.config.SyncEncryptionPassphrase != "" {
		//Asterisk indicates user wants to provide passphrase on terminal
		if strings.HasPrefix(s.config.SyncEncryptionPassphrase, "*") {
			encryptionPassphrase = passphraseInput()
			//Else a full passphrase was provided in config file
		} else {
			encryptionPassphrase = s.config.SyncEncryptionPassphrase
		}

		tmpfile, err := ioutil.TempFile("", "temp_sync_backlog.json")
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(tmpfile.Name()) // clean up
		origSyncFilepath = syncFilepath
		syncFilepath = tmpfile.Name()
		//println("origSyncFilepath: ", origSyncFilepath)
		//println("syncFilepath: ", syncFilepath)
		//decrypt remote file, then write plain text to temp file location
		decryptFile(origSyncFilepath, syncFilepath, encryptionPassphrase)
	}

	store := s.store
	var err error
	var todos []*Todo
	todos, err = store.LoadPending()
	if err != nil {
		println("Error reading pending todos for sync job")
		return err
	}
	s.Local.Load(todos)
	todos, err = store.LoadArchived()
	if err != nil {
		println("Error reading archived todos for sync job")
		return err
	}
	s.Local.Load(todos)
	todos, err = store.LoadBacklog(store.GetBacklogFilepath())
	if err != nil {
		println("Error reading local backlog todos for sync job")
		return err
	}
	backlogCount := 0
	s.Backlog.Load(todos)
	if len(s.Backlog.Data) > 0 {
		if s.Backlog.Data[0].Status == "Checkpoint" {
		s.Checkpoint = s.Backlog.Data[0]
		backlogCount = len(todos) - 1
		} else {
			s.Checkpoint = NewTodo()
			s.Checkpoint.Status = "Checkpoint"
			s.Checkpoint.IsModified = true
			backlogCount = len(todos)
		}
	} else {
		s.Checkpoint = NewTodo()
		s.Checkpoint.Status = "Checkpoint"
		s.Checkpoint.IsModified = true
	}
	todos, err = store.LoadBacklog(syncFilepath)
	if err != nil {
		todos = []*Todo{}
	}
	todos = s.newSinceLastSync(todos, s.Checkpoint)
	s.Remote.Load(todos)

	//Synchronize the remote and local data
	s.syncRemoteChanges()

	//Set new checkpoint to be used in backlog and remote backlog
	newCheckpoint := NewTodo()
	newCheckpoint.Status = "Checkpoint"
	newCheckpoint.ModifiedDate = time.Now().Format(time.RFC3339)
	newCheckpoint.IsModified = true

	if len(s.Backlog.Data) > 0 && s.Backlog.Data[0].Status == "Checkpoint" {
		s.Backlog.Data = s.Backlog.Data[1:] //remove starting checkpoint before updating remote file
	}
	s.Backlog.Data = append(s.Backlog.Data, newCheckpoint)
	store.AppendBacklog(syncFilepath, s.Backlog.Data)

	//Re-load Remote Backlog and remove prior checkpoint
	s.RemovePriorCheckpointFromSyncFile(syncFilepath)

	//If encrypting, read temp file, re-encrypt and write to orig sync file location
	if encryptionPassphrase != "" {
		encryptFile(syncFilepath, origSyncFilepath, encryptionPassphrase)
		store.DeleteBacklog(syncFilepath) //delete temporary file
	}

	//Delete existing local backlog file
	store.DeleteBacklog(store.GetBacklogFilepath())

	//Ensure none of the todos has IsModified == true so none end up in the backlog
	for _, todo := range s.Local.Data {
		todo.IsModified = false
	}
	//Re-assign all the Ids based on Uuid sort order so that two repos
	//that are synced to the same point will have the same ids for same task
	s.Local.ReassignAllIds()
	//Add newCheckpoint so it is the only entry in the new backlog file
	s.Local.Data = append(s.Local.Data, newCheckpoint)
	//Save will write todos into pending, archived and backlog files
	store.Save(s.Local.Data)

	//Print stats about the sync
	//No. of Todos added, modified, deleted
	fmt.Println("Sync completed.")
	fmt.Println("\tUploaded: ", backlogCount)
	fmt.Println("\tAdded: ", len(s.addedTodos))
	fmt.Println("\tModified: ", len(s.modifiedTodos))
	fmt.Println("\tDeleted: ", len(s.deletedTodos))

	//If verbose flag is set, also print the actual todos added, modified and deleted
	if verbose {
		printer := NewScreenPrinter()
		report, ok := s.config.GetReport("default")
		if !ok {
			fmt.Println("ERROR getting default report from config. Can't print Todos")
			os.Exit(1)
		}
		if len(s.addedTodos) > 0 {
			fmt.Println("Added:")
			printer.PrintReport(report, s.addedTodos)
		}
		if len(s.modifiedTodos) > 0 {
			fmt.Println("Modified:")
			printer.PrintReport(report, s.modifiedTodos)
		}
		if len(s.deletedTodos) > 0 {
			fmt.Println("Deleted:")
			printer.PrintReport(report, s.deletedTodos)
		}
	}
	return nil
}

func (s *TodoSync) RemovePriorCheckpointFromSyncFile(syncFilepath string) {
	todos, err := s.store.LoadBacklog(syncFilepath)
	if err != nil {
		todos = []*Todo{}
	}
	todos = s.consolidateBacklog(todos)
	idx := 0
	for _, todo := range todos {

		if todo.Uuid == s.Checkpoint.Uuid {
			//println("Removing prior Uuid: ", s.Checkpoint.Uuid)
			if idx < len(todos)-1 {
				todos = append(todos[0:idx], todos[idx+1:]...)
			} else {
				todos = todos[0:idx]
			}
		} else {
			idx++
		}
	}
	//Delete existing local backlog file
	s.store.DeleteBacklog(syncFilepath)
	//Write new backlog file
	s.store.AppendBacklog(syncFilepath, todos)
}

func (s *TodoSync) newSinceLastSync(todos []*Todo, checkpoint *Todo) []*Todo {
	ret := []*Todo{}
	postCheckpoint := false
	if checkpoint.ModifiedDate == "" {
		postCheckpoint = true
	} else {
		hasMatchingCheckpoint := false
		for _, todo := range todos {
			if todo.Uuid == checkpoint.Uuid {
				hasMatchingCheckpoint = true
				break
			}
		}
		if !hasMatchingCheckpoint {
			postCheckpoint = true
		}
	}
	for _, todo := range todos {
		if !postCheckpoint {
			if todo.Uuid == checkpoint.Uuid {
				postCheckpoint = true
			}
			continue
		}
		ret = append(ret, todo)
	}
	return ret
}

func (s *TodoSync) consolidateBacklog(todos []*Todo) []*Todo {
	ret := []*Todo{}
	m := map[string]int{} //Uuid to index
	for _, todo := range todos {
		if _, ok := m[todo.Uuid]; ok {
			idx := m[todo.Uuid]
			ret = append(ret[0:idx], ret[idx+1:]...) //remove earlier entry
		}
		ret = append(ret, todo)
		m[todo.Uuid] = len(ret) - 1 //record location of entry
	}
	return ret
}

func (s *TodoSync) syncRemoteChanges() {
	for _, remoteTodo := range s.Remote.Data {
		if remoteTodo.Status == "Checkpoint" {
			continue
		}
		isMatched := false
		for _, localTodo := range s.Local.Data {
			if localTodo.Uuid == remoteTodo.Uuid {
				isMatched = true
				//May not need return value here. LocalTodo is updated in s.Local.Data
				localTodo = s.diffTodos(localTodo, remoteTodo)
				if localTodo.Status == "Deleted" {
					s.deletedTodos = append(s.deletedTodos, localTodo)
				} else {
					s.modifiedTodos = append(s.modifiedTodos, localTodo)
				}
			}
		}
		if !isMatched {
			if remoteTodo.Status == "Deleted" {
				s.deletedTodos = append(s.deletedTodos, remoteTodo)
			} else {
				remoteTodo.Id = s.Local.NextId()
				s.Local.Data = append(s.Local.Data, remoteTodo)
				s.addedTodos = append(s.addedTodos, remoteTodo)
			}
		}
	}
}

func (s *TodoSync) diffTodos(local *Todo, remote *Todo) *Todo {

	localTime := s.getModifiedTime(local)
	remoteTime := s.getModifiedTime(remote)

	if remoteTime.After(localTime) {
		local.Subject = remote.Subject
		local.Priority = remote.Priority
		local.ModifiedDate = remote.ModifiedDate
		local.Wait = remote.Wait
		local.Until = remote.Until
		local.Due = remote.Due
		local.Completed = remote.Completed
		local.CompletedDate = remote.CompletedDate
		local.Status = remote.Status
		local.Notes = remote.Notes
		//Determine if adding or removing projects from local
		//and invoke todolist.AddProject or todolist.RemoveProject, which
		//will ensure the ordinals are updated.
		toRemove := s.inSliceOneNotSliceTwo(local.Projects, remote.Projects)
		toAdd := s.inSliceOneNotSliceTwo(remote.Projects, local.Projects)
		for _, p := range toRemove {
			s.Local.RemoveProject(p, local)
		}
		for _, p := range toAdd {
			s.Local.AddProject(p, local)
		}
		toRemove = s.inSliceOneNotSliceTwo(local.Contexts, remote.Contexts)
		toAdd = s.inSliceOneNotSliceTwo(remote.Contexts, local.Contexts)
		for _, c := range toRemove {
			s.Local.RemoveContext(c, local)
		}
		for _, c := range toAdd {
			s.Local.AddContext(c, local)
		}
		local.Ordinals = remote.Ordinals //Do this and above add/remove or just copy remote?
	}

	//If remoteTime is older than localTime, but still newer than the last sync time
	//then what? Assume local with more recent timestamp is most up to date
	//on all elements. Not likely to be true, but simple and reasonbly easy rule
	//to follow that should do the right thing most of the time.
	return local
}

func (s *TodoSync) inSliceOneNotSliceTwo(s1, s2 []string) []string {
	// difference returns the elements in s1 that aren't in s2
	ms2 := map[string]bool{} //map of slice 2 elements
	for _, x := range s2 {
		ms2[x] = true
	}
	res := []string{} //result slice to contain s1 elements not in s2
	for _, x := range s1 {
		if _, ok := ms2[x]; !ok {
			res = append(res, x)
		}
	}
	return res
}

func (s *TodoSync) getModifiedTime(todo *Todo) time.Time {
	if len(todo.ModifiedDate) > 0 {
		modTime, rerr := time.Parse(time.RFC3339, todo.ModifiedDate)
		if rerr != nil {
			createTime, _ := time.Parse(time.RFC3339, todo.CreatedDate)
			return createTime
		}
		return modTime
	}
	return time.Now()
}

func (s *TodoSync) getStatus(todo *Todo) int {
	if todo.Status == "Pending" {
		return 0
	} else if todo.Status == "Archived" {
		return 1
	} else if todo.Status == "Deleted" {
		return 2
	}
	return 0
}

/**
TODO - Expand the following logic

Exec any external script to pull from remote site to local "remote" file
Do the sync
Exec any external script to push local "remote" file to remote location

*/
func (s *TodoSync) Run(script string) bool {
	c := exec.Command(script)

	if err := c.Run(); err != nil {
		fmt.Println("Error: ", err)
		return false
	}
	return true
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func decrypt(data []byte, passphrase string) []byte {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

func encryptFile(srcFilename string, dstFilename string, passphrase string) bool {
	data, _ := ioutil.ReadFile(srcFilename)
	f, _ := os.Create(dstFilename)
	defer f.Close()
	_, err := f.Write(encrypt(data, passphrase))
	return err != nil
}

func decryptFile(srcFilename string, dstFilename string, passphrase string) bool {
	data, _ := ioutil.ReadFile(srcFilename)
	if len(data) == 0 {
		return true
	}
	decrypted := decrypt(data, passphrase)
	err := ioutil.WriteFile(dstFilename, decrypted, os.FileMode(os.O_RDWR))
	return err != nil
}

func passphraseInput() string {
	fmt.Print("Enter Password: ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	//if err == nil {
	//	fmt.Println("\nPassword typed: " + string(bytePassword))
	//}
	password := string(bytePassword)
	fmt.Println("")
	return strings.TrimSpace(password)
}

/*
func main() {
	fmt.Println("Starting the application...")
	ciphertext := encrypt([]byte("Hello World"), "password")
	fmt.Printf("Encrypted: %x\n", ciphertext)
	plaintext := decrypt(ciphertext, "password")
	fmt.Printf("Decrypted: %s\n", plaintext)
	encryptFile("sample.txt", []byte("Hello World"), "password1")
	fmt.Println(string(decryptFile("sample.txt", "password1")))
}
*/
