package courses

import "os"

type Course struct {
	Id             int    `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Price          int    `json:"price"`
	PathToCloudDir string ``
}

type UserCourses struct {
	Id       int
	UserId   int
	CourseId int
}

type CourseItem struct {
	Id    int     `json:"id"`
	Title string  `json:"title"`
	Video os.File `json:"video"`
	Photo os.File `json:"photo"`
	Pdf   os.File `json:"pdf"`
}

type ListItem struct {
	Id       int
	UserId   int
	CourseId int
}
