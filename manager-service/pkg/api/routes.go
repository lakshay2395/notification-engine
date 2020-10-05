package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/lakshay2395/notification-engine/manager-service/pkg/db"
	"github.com/lakshay2395/notification-engine/manager-service/pkg/queue"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
)

type API struct {
	router *mux.Router
	dbClient db.DB
	workerQueue queue.Queue
}

func (a *API) InitRoutes(port string) error{
	a.router.HandleFunc("/trigger/task/{taskId}", a.triggerTaskHandler).Methods("PUT")
	a.router.HandleFunc("/trigger-shards", a.triggerShardHandler).Methods("PUT")
	a.router.HandleFunc("/tasks", a.addTasksHandler).Methods("POST")
	a.router.HandleFunc("/delete/task/{taskId}", a.deleteTaskHandler).Methods("DELETE")
	return http.ListenAndServe(":"+port,a.router)
}

func (a *API) addTasksHandler(w http.ResponseWriter, r *http.Request) {
	var tasks []db.Task
	request,err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(request,&tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err =  a.dbClient.AddTasks(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Tasks Added Successfully"))
}

func (a *API) deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId := vars["taskId"]
	uid,err := uuid.FromString(taskId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = a.dbClient.DeleteTask(uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Task Deleted Successfully"))
}

func (a *API) triggerTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId := vars["taskId"]
	uid,err := uuid.FromString(taskId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	task,err := a.dbClient.GetTaskById(uid)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if task.Status != db.UNPICKED {
		http.Error(w,"Task already picked",http.StatusBadRequest)
		return
	}
	err = a.updateStatusAndPublishTaskToWorkerQueue(task)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Task triggered Successfully"))
}

func (a *API) triggerShardHandler(w http.ResponseWriter, r *http.Request) {
	data := TriggerShardRequest{}
	request,err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(request,&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	shards := data.Shards
	for _,shard := range shards {
		tasks,err := a.dbClient.GetUnpickedTasksByShardId(shard)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _,task := range tasks {
			err = a.updateStatusAndPublishTaskToWorkerQueue(&task)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
	w.Write([]byte("Shards triggered Successfully"))
}

func (a *API) updateStatusAndPublishTaskToWorkerQueue(task *db.Task) error {
	err := a.dbClient.UpdateTaskStatus(task.Id,db.PENDING)
	if err != nil {
		return err
	}
	data,err := json.Marshal(task)
	if err != nil {
		return err
	}
	err = a.workerQueue.Publish(string(data))
	if err != nil {
		return err
	}
	return nil
}

func Init(dbClient db.DB,workerQueue queue.Queue) *API{
	return &API{
		router: mux.NewRouter(),
		dbClient: dbClient,
		workerQueue: workerQueue,
	}
}