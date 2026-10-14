package domain

type Job struct {
	ID, RequestID, InputHash, ProjectID, NodeID, ChannelID, Model string
	Kind Kind
	Prompt, Status, Text, AssetID, ErrorCode, CreatedAt, UpdatedAt string
}
func (j Job) Pending() bool { return j.Status == "queued" || j.Status == "running" }
type GenerationRequest struct { Model string; Kind Kind; Prompt string }
type GenerationResult struct { Text string; Media []byte; MediaName string }
type Asset struct { ID, Name, MIMEType string; Bytes int64; CreatedAt string }
func (a Asset) Kind() Kind {
	if len(a.MIMEType) >= 6 && a.MIMEType[:6] == "image/" { return Image }
	if len(a.MIMEType) >= 6 && a.MIMEType[:6] == "video/" { return Video }
	return Audio
}
type LibraryEntry struct { ID, Kind, Name, Content, CreatedAt, UpdatedAt string }
