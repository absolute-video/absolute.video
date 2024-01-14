package transcode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type Stream struct {
    Height int `json:"height"`
    Width  int `json:"width"`
}

type VideoData struct {
    Streams []Stream `json:"streams"`
}

type TranscodeSize struct {
	Height int
	Width int
}

type TranscodedVideoPath struct {
	Path string
	Size TranscodeSize
}

type UploadedVideoPath struct {
	Path string
	Size TranscodeSize
}


func GetTranscodaleSizes(height int, width int) ([]TranscodeSize) {
	aspectRatio := float32(width) / float32(height)
	var currentMeasurableSize int
	if  height < width {
		currentMeasurableSize = height
	} else {
		currentMeasurableSize = width
	}
	supportedSizes  := [5]int{240,360,480,720,1080}

	smallerSizes := []int{}
	for _,size := range supportedSizes {
		if size < currentMeasurableSize {
		smallerSizes = append(smallerSizes, size)
		}
	}


	// then we will create the dimmentions for the smaller sizes
	allTranscodableSizes := []TranscodeSize{}
	for _,size := range smallerSizes {
		var transcodeSize TranscodeSize
		if  height < width{
				height := size
				width := int(float32(height) * aspectRatio)
				transcodeSize = TranscodeSize{height,width}
			}else{
				width := size
				height := int(float32(width) / aspectRatio)
				transcodeSize = TranscodeSize{height,width}
		}
		allTranscodableSizes = append(allTranscodableSizes, transcodeSize)
	}

	return allTranscodableSizes
}

func GetVideoInfo(path string) (VideoData, error) {
	info, err := ffmpeg.Probe(path)
	if err != nil {
		return VideoData{}, err
	}

	fmt.Println(info)
	var data VideoData
	err = json.Unmarshal([]byte(info), &data)

	return data, nil
}

func TransCodeEachVideo(path string,size  TranscodeSize,transcodedVideoPath chan TranscodedVideoPath ) {
	widht := (size.Width/2)*2 // make sure the width is even
	height := (size.Height/2)*2 // make sure the height is even
	ontputPath := fmt.Sprintf("tmp/converted/output_%d_%d.mp4",widht,height)

	err := ffmpeg.Input(path).
		Output(ontputPath,
		ffmpeg.KwArgs{"vf": fmt.Sprintf("scale=w=%d:h=%d",widht,height)}).
		OverWriteOutput().ErrorToStdOut().Run()
			
	if err != nil {
		transcodedVideoPath <-  TranscodedVideoPath{"",size}
	}
	transcodedVideoPath <- TranscodedVideoPath{ontputPath,size}
}

func TranscodeVideo(path string, transcodableSizes []TranscodeSize) ([]TranscodedVideoPath, error) {
    transcodedVideoPathChans := make(chan TranscodedVideoPath)
    var wg sync.WaitGroup // Using WaitGroup to wait for all goroutines to finish

    // Start a goroutine for each transcode operation
    for _, size := range transcodableSizes {
        wg.Add(1) // Increment the WaitGroup counter
        go func(size TranscodeSize) {
            defer wg.Done() // Decrement the counter when the goroutine completes
            TransCodeEachVideo(path, size, transcodedVideoPathChans)
        }(size)
    }

    // Start a goroutine to close the channel once all transcoding is done
    go func() {
        wg.Wait() // Wait for all transcoding goroutines to complete
        close(transcodedVideoPathChans) // Close the channel
    }()

    // Collecting results from the channel
    var transcodedVideoPaths []TranscodedVideoPath
    for videoPath := range transcodedVideoPathChans {
        transcodedVideoPaths = append(transcodedVideoPaths, videoPath)
    }

    return transcodedVideoPaths, nil
}

func UploadEachVideoToGoogleStorage(path string,object string)( string, error ){
	ctx:= context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Println(err)
		return "", err
	}


	defer client.Close()

	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer f.Close()
	b			 := client.Bucket("absolute-video").Object(object) // where object is the name of the file

	ctx,cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	wc := b.NewWriter(ctx)
	
	if _, err = io.Copy(wc, f); err != nil {
		fmt.Println(err)
		return "", err
	}
	if err := wc.Close(); err != nil {
		fmt.Println(err)
		return "", err
	}

	attrs, err := b.Attrs(ctx)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return attrs.MediaLink, nil
	
}

func UploadVideosToGoogleStorage(videos []TranscodedVideoPath) []UploadedVideoPath {
	var uploadedVideoPaths  []UploadedVideoPath
	for _,video := range videos {
		uploadedVideoPath,err := UploadEachVideoToGoogleStorage(video.Path,fmt.Sprintf("transcoded/%d_%d.mp4",video.Size.Width,video.Size.Height))
		
		if err != nil {
			uploadedVideoPaths = append(uploadedVideoPaths, UploadedVideoPath{"",video.Size})
			continue
		}
		uploadedVideoPaths = append(uploadedVideoPaths, UploadedVideoPath{uploadedVideoPath,video.Size})
	}
	return uploadedVideoPaths
}