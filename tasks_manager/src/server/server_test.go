package server

import (
	pb "tasksmanager/proto"
	"tasksmanager/src/database"
	"testing"
)

func TestDataToProto(t *testing.T) {
	in := &database.TaskData{ID: 1, Author: "1", Title: "T1", Content: "C1"}
	target := &pb.Task{Id: 1, Author: "1", Title: "T1", Content: "C1"}

	out := DataToProto(in)
	if out.Id != target.Id ||
		out.Author != target.Author ||
		out.Title != target.Title ||
		out.Content != target.Content {
		t.Errorf("expected: %#v; got: %#v", target, out)
	}
}

func TestDataToTasks(t *testing.T) {
	in := []database.TaskData{
		{ID: 1, Author: "1", Title: "T1", Content: "C1"},
		{ID: 2, Author: "2", Title: "T2", Content: "C2"},
	}
	target := []*pb.Task{
		{Id: 1, Author: "1", Title: "T1", Content: "C1"},
		{Id: 2, Author: "2", Title: "T2", Content: "C2"},
	}

	out := dataToTasks(in)
	for i, task := range out {
		if task.Id != target[i].Id ||
			task.Author != target[i].Author ||
			task.Title != target[i].Title ||
			task.Content != target[i].Content {
			t.Errorf("expected: %#v; got: %#v", target[i], task)
		}
	}
}
