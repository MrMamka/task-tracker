package server

import (
	pb "statistics/proto/statistic"
	"statistics/src/database"
	"testing"
)

func TestExtractTopAuthors(t *testing.T) {
	in := map[string]int64{
		"a": 1,
		"b": 3,
		"c": 100,
		"d": 10,
	}
	target := []*pb.Author{
		{Login: "c", Likes: 100},
		{Login: "d", Likes: 10},
		{Login: "b", Likes: 3},
	}

	out := extractTopAuthors(in)
	for i, author := range out {
		if author.Login != target[i].Login || author.Likes != target[i].Likes {
			t.Errorf("expected: %#v; got: %#v", target[i], author)
		}
	}
}

func TestExtractTopTasks(t *testing.T) {
	in := []database.TaskIDCount{
		{TaskID: 1, Count: 3},
		{TaskID: 2, Count: 2},
		{TaskID: 3, Count: 1},
	}
	authors := map[uint32]string{
		1: "1",
		2: "2",
		3: "3",
	}
	target := []*pb.Task{
		{Id: 1, LikesOrViews: 3, Author: "1"},
		{Id: 2, LikesOrViews: 2, Author: "2"},
		{Id: 3, LikesOrViews: 1, Author: "3"},
	}

	out := extractTopTasks(in, authors)
	for i, author := range out {
		if author.Id != target[i].Id ||
			author.LikesOrViews != target[i].LikesOrViews ||
			author.Author != target[i].Author {
			t.Errorf("expected: %#v; got: %#v", target[i], author)
		}
	}
}
