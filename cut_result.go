package ad_cutter

type CutResult struct {
	Video        string // 视频文件绝对路径
	IsAd         bool   // 是否包含广告
	IsCutted     bool   // 是否已完成裁剪
	ErrorMessage string // 异常
}
