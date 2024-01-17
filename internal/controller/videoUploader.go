package videoUploader

import (
	"net/http"

	"absolute.video/internal/transcode"
	"github.com/gin-gonic/gin"
)



	
func VideoHandler(c *gin.Context) {

	file, err := c.FormFile("file")
	
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	

	pathName := "tmp/uploaded/"+file.Filename
	// fmt.Println(pathName)
	err = c.SaveUploadedFile(file,pathName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	data,err := transcode.GetVideoInfo(pathName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	height,width := data.Streams[0].Height,data.Streams[0].Width

	transcodableSizes := transcode.GetTranscodaleSizes(height,width)

	// upload original video
	// transcode video
	paths,err := transcode.TranscodeVideo(pathName,transcodableSizes)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	
	rsp:=transcode.UploadVideosToGoogleStorage(paths)
	
	c.JSON(http.StatusOK, gin.H{"rsp":rsp,"transcodableSizes":transcodableSizes})
}