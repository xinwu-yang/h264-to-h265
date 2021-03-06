package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"

	"go.uber.org/zap"
)

const BufferSize = 32768

type Format struct {
	FileName       string  `json:"filename"`
	FormatLongName string  `json:"format_long_name"`
	Duration       float32 `json:"duration,string"`
	Size           int64   `json:"size,string"`
	BitRate        int64   `json:"bit_rate,string"`
}

type Stream struct {
	Index     int    `json:"index"`
	CodecType string `json:"codec_type"`
	CodecName string `json:"codec_name"`
}

type VideoSteam struct {
	Stream
	PixelFormat string `json:"pix_fmt"`
}

type AudioSteam struct {
	Stream
	Channels int `json:"channels"`
}

type MediaInfo struct {
	Format     Format
	VideoSteam VideoSteam
	AudioSteam AudioSteam
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	baseDir := "D:\\demo-video"
	f, err := os.OpenFile(baseDir, os.O_RDONLY, os.ModeDir)
	if err != nil {
		sugar.Errorf("open dir has error", "err", err.Error())
	}
	defer f.Close()
	dirs, _ := f.ReadDir(-1)
	for _, dir := range dirs {
		if !dir.IsDir() {
			sugar.Infof("开始处理文件：", dir.Name())
			fileName := baseDir + "\\" + dir.Name()
			sugar.Infof("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", fileName)
			cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", fileName)
			ffprobeOut, _ := cmd.StdoutPipe()
			cmd.Start()
			var bt bytes.Buffer
			for {
				readData := make([]byte, BufferSize)
				i, _ := ffprobeOut.Read(readData)
				if i > 0 {
					bt.Write(readData[:i])
				} else {
					// 读取完输出后解析json
					//videoInfoJson := bt.String()
					//fmt.Println(videoInfoJson)
					format := Format{}
					videoSteam := VideoSteam{}
					audioSteam := AudioSteam{}
					mediaInfo := MediaInfo{}
					jsonBytes := bt.Bytes()
					var data map[string]interface{}
					err := json.Unmarshal(jsonBytes[:bt.Len()], &data)
					if err != nil {
						sugar.Errorf(err.Error())
						return
					}
					formatBytes, _ := json.Marshal(data["format"])
					json.Unmarshal(formatBytes, &format)

					streams := data["streams"]
					streamsBytes, _ := json.Marshal(streams)
					var streamData []map[string]interface{}
					json.Unmarshal(streamsBytes, &streamData)

					for _, stream := range streamData {
						streamBytes, _ := json.Marshal(stream)
						if stream["codec_type"] == "video" {
							json.Unmarshal(streamBytes, &videoSteam)
						} else if stream["codec_type"] == "audio" {
							json.Unmarshal(streamBytes, &audioSteam)
						}
					}
					mediaInfo.Format = format
					mediaInfo.VideoSteam = videoSteam
					mediaInfo.AudioSteam = audioSteam
					handleVideoCodec, handleVideoPixFmt, handleAudioCodec, handleAudioChannels := false, false, false, false

					// 根据参数判断是否处理视频
					if videoSteam.CodecType == "video" && videoSteam.CodecName != "hevc" {
						handleVideoCodec = true
					}
					if videoSteam.CodecType == "video" && videoSteam.PixelFormat != "yuv420p" {
						handleVideoPixFmt = true
					}
					if audioSteam.CodecType == "audio" && audioSteam.CodecName != "aac" {
						handleAudioCodec = true
					}
					if audioSteam.CodecType == "audio" && audioSteam.Channels != 2 {
						handleAudioChannels = true
					}
					// 开始处理视频
					sugar.Infof("是否处理视频编码：%f\n", handleVideoCodec)
					sugar.Infof("是否处理视频像素格式：%f\n", handleVideoPixFmt)
					sugar.Infof("是否处理音频编码：%f\n", handleAudioCodec)
					sugar.Infof("是否处理音频声道数：%f\n", handleAudioChannels)
					if handleVideoCodec {
						handleVideo(fileName, mediaInfo, handleVideoCodec, handleVideoPixFmt, handleAudioCodec, handleAudioChannels)
					}
					break
				}
			}
			ffprobeOut.Close()
		}
	}
}

/*处理视频*/
func handleVideo(fileName string, mediaInfo MediaInfo, handleVideoCodec bool, handleVideoPixFmt bool, handleAudioCodec bool, handleAudioChannels bool) {
	//ffmpegCmdArray := []string{}
	//ffmpegCmd := exec.Command("ffmpeg", "-i", fileName)
}
