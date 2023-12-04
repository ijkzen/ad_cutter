package ad_cutter

import (
	"os"
)

func Cut(video string) CutResult {
	result := CutResult{}
	result.Video = video
	output, err := generateRawData(video)
	if err != nil {
		result.ErrorMessage = err.Error()
		return result
	} else {
		outputFile, err := os.Open(output)
		if err != nil {
			result.ErrorMessage = err.Error()
			return result
		}

		stat, err := outputFile.Stat()
		if err != nil {
			result.ErrorMessage = err.Error()
			return result
		}

		duration := int(stat.Size() / 8000)
		pcmData := make([]byte, stat.Size())
		if _, err := outputFile.Read(pcmData); err != nil {
			result.ErrorMessage = err.Error()
			return result
		}

		os.Remove(output)
		resolution := 1000
		sliceWidth := stat.Size() / int64(resolution)

		i := int64(0)
		resample := make([]int, 0)
		for i+sliceWidth < stat.Size() {
			resample = append(resample, int(pcmData[i])-127)
			i += sliceWidth
		}

		zeroIndexCandidates := make([]int, 0)

		for index, value := range resample {
			if value == 0 {
				zeroIndexCandidates = append(zeroIndexCandidates, index)
			}
		}

		// 初次过滤零点，零点左边至少大于两倍右边的零点保留
		zeroAccept := make([]int, 0)
		for _, zeroIndex := range zeroIndexCandidates {
			leftSlice := resample[0:zeroIndex]
			rightSlice := resample[zeroIndex+1:]

			if zeroIndex > resolution-zeroIndex {
				continue
			}

			leftSum := int64(0)
			for _, value := range leftSlice {
				if value > 0 {
					leftSum += int64(value)
				}
			}

			rightSum := int64(0)
			for _, value := range rightSlice {
				if value > 0 {
					rightSum += int64(value)
				}
			}

			if leftSum > 2*rightSum {
				zeroAccept = append(zeroAccept, zeroIndex)
			}
		}

		cutIndex := 0
		maxRate := 0
		if len(zeroAccept) > 0 {
			result.IsAd = true

			for _, zeroIndex := range zeroAccept {
				leftSlice := resample[0:zeroIndex]
				rightSlice := resample[zeroIndex+1:]

				leftCount := float64(0)
				rightCount := float64(0)

				for _, value := range leftSlice {
					tmp := 0
					if value < 0 {
						tmp = -value
					} else {
						tmp = value
					}

					if tmp > 20 {
						leftCount++
					}
				}

				for _, value := range rightSlice {
					tmp := 0
					if value < 0 {
						tmp = -value
					} else {
						tmp = value
					}

					if tmp > 20 {
						rightCount++
					}
				}

				leftDensity := leftCount / float64(zeroIndex)
				rightDensity := rightCount / float64(resolution-zeroIndex-1)

				if int(leftDensity/rightDensity) > maxRate {
					cutIndex = zeroIndex
					maxRate = int(leftDensity / rightDensity)
				}

			}

			second := int(float64(duration) * (float64(cutIndex) / float64(resolution)))
			_, err := cutMoive(video, second)
			if err != nil {
				result.IsCutted = false
				result.ErrorMessage = err.Error()
			} else {
				result.IsCutted = true
			}
		} else {
			result.IsAd = false
		}

		return result
	}
}
