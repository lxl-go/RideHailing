package response

import "ride-hailing/admin-server/config"

type SysConfigResponse struct {
	Config config.Server `json:"config"`
}
