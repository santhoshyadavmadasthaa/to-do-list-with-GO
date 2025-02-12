package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/mergestat/timediff"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

type Task struct {
	ID        int
	Desc      string
	CreatedAt time.Time
	Complete  bool
}

const filePath = "tasks.csv"

func main() {
	rootCmd := &cobra.Command{Use: "tasks"}
	rootCmd.AddCommand(addCmd, listCmd, completeCmd, deleteCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

var addCmd = &cobra.Command{
	Use:   "add [description]",
	Short: "Add a new task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		desc := args[0]
		task := Task{Desc: desc, CreatedAt: time.Now(), Complete: false}
		addTask(task)
		fmt.Println("Task added:", desc)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		listTasks(all)
	},
}

var completeCmd = &cobra.Command{
	Use:   "complete [task ID]",
	Short: "Mark a task as complete",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := strconv.Atoi(args[0])
		completeTask(id)
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [task ID]",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := strconv.Atoi(args[0])
		deleteTask(id)
	},
}

func addTask(task Task) {
	tasks := readTasks()
	task.ID = len(tasks) + 1
	tasks = append(tasks, task)
	writeTasks(tasks)
}

func listTasks(all bool) {
	tasks := readTasks()
	fmt.Printf("ID\tTask\t\t\tCreated\t\tDone\n")
	for _, task := range tasks {
		if all || !task.Complete {
			fmt.Printf("%d\t%s\t\t%s\t%t\n", task.ID, task.Desc, timediff.TimeDiff(task.CreatedAt), task.Complete)
		}
	}
}

func completeTask(id int) {
	tasks := readTasks()
	for i, task := range tasks {
		if task.ID == id {
			tasks[i].Complete = true
			writeTasks(tasks)
			fmt.Println("Task marked as complete:", task.Desc)
			return
		}
	}
	fmt.Println("Task not found.")
}

func deleteTask(id int) {
	tasks := readTasks()
	var updatedTasks []Task
	for _, task := range tasks {
		if task.ID != id {
			updatedTasks = append(updatedTasks, task)
		}
	}
	writeTasks(updatedTasks)
	fmt.Println("Task deleted.")
}

func readTasks() []Task {
	f, err := loadFile(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer closeFile(f)

	reader := csv.NewReader(f)
	records, _ := reader.ReadAll()
	var tasks []Task
	for _, record := range records {
		id, _ := strconv.Atoi(record[0])
		createdAt, _ := time.Parse(time.RFC3339, record[2])
		complete, _ := strconv.ParseBool(record[3])
		tasks = append(tasks, Task{ID: id, Desc: record[1], CreatedAt: createdAt, Complete: complete})
	}
	return tasks
}

func writeTasks(tasks []Task) {
	f, err := loadFile(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer closeFile(f)

	writer := csv.NewWriter(f)
	defer writer.Flush()

	for _, task := range tasks {
		writer.Write([]string{
			strconv.Itoa(task.ID),
			task.Desc,
			task.CreatedAt.Format(time.RFC3339),
			strconv.FormatBool(task.Complete),
		})
	}
}

func loadFile(filepath string) (*os.File, error) {
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for reading")
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, err
	}
	return f, nil
}

func closeFile(f *os.File) error {
	unix.Flock(int(f.Fd()), unix.LOCK_UN)
	return f.Close()
}
