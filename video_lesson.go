package courses

type VideoLesson struct {
	Id              uint     `db:"id"`
	Title           string   `json:"title" binding:"required" db:"title"`
	Description     string   `json:"description" binding:"required" db:"description"`
	VideoName       string   `json:"video_name" binding:"required" db:"video_name"`
	AdditionalFiles []VLItem `json:"additional_files" binding:"required"`
	Price           int      `json:"price" binding:"required" db:"price"`
	PathToCloudDir  string   `db:"path_to_cloud_dir"`
}

type VLItem struct {
	Id          uint   `db:"id"`
	Data        []byte `json:"data" binding:"required"`
	ContentType string `json:"content_type" db:"content-type"`
	FileName    string `json:"file_name"`
	PathToCloud string `db:"path_to_cloud"`
}

type UsersVideoLessons struct {
	Id            uint
	UserId        uint
	VideoLessonId uint
}

type VideoLessonsItems struct {
	Id            uint
	VideoLessonId uint
	VLItemId      uint
}
