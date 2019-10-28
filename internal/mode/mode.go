package mode

import (
	"context"

	"go.uber.org/zap"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
)

// OperationModeFunc represents operation mode function.
type OperationModeFunc func(context.Context, *zap.SugaredLogger, clients.Client, *entities.Job) entities.Result

const (
	// ChangePasswordMode is change password operation name.
	ChangePasswordMode = "ChangePassword"
	// CustomSSHMode is custom SSH job operation name.
	CustomSSHMode = "CustomSSH"
	// CustomAPIMode is custom Mikrotik secure API job operation name.
	CustomAPIMode = "CustomAPI"
	// InitSecureAPIMode is initialize secure API job operation name.
	InitSecureAPIMode = "InitSecureAPI"
	// InitPublicKeySSHMode is initialize public key SSH authentication job operation name.
	InitPublicKeySSHMode = "InitPublicKeySSH"
	// CheckMTbulkVersionMode is check MT-bulk version job operation name.
	CheckMTbulkVersionMode = "CheckMTbulkVersion"
	// SFTPMode is SFTP job operation name.
	SFTPMode = "SFTP"
	// SystemBackupMode is system backup job operation name.
	SystemBackupMode = "SystemBackup"
	// SecurityAuditMode is name of a job performing security audit of device.
	SecurityAuditMode = "SecurityAudit"
)
