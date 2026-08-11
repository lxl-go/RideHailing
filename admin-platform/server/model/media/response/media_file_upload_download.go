package response

import "ride-hailing/admin-server/model/media"

type FileUploadAndDownloadResponse struct {
	File media.FileUploadAndDownload `json:"file"`
}
