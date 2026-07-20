package extensions

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrRestoreTargetNotStopped = errors.New("extension_restore_target_not_stopped")
	ErrRestoreTargetNotEmpty   = errors.New("extension_restore_target_not_empty")
	ErrBackupCodecUnsupported  = errors.New("extension_backup_codec_unsupported")
	ErrRestoreFailed           = errors.New("extension_restore_failed")
)

type RestoreTarget struct {
	Stopped bool
	Empty   bool
	Failed  bool
}

type RestoreBindingInput struct {
	BindingID   string
	CodecID     string
	CodecSHA256 string
	Entries     [][]byte
}

type BackupRestorePort interface {
	RestoreBinding(context.Context, BackupBinding, BackupCodec, RestoreBindingInput) error
	ValidateBinding(context.Context, BackupBinding) error
}

type BackupCoordinator struct{}

func (BackupCoordinator) Restore(ctx context.Context, target *RestoreTarget, plan BackupPlan, inputs []RestoreBindingInput, port BackupRestorePort) error {
	if target == nil || !target.Stopped {
		return ErrRestoreTargetNotStopped
	}
	if !target.Empty {
		return ErrRestoreTargetNotEmpty
	}
	if target.Failed || port == nil {
		return ErrRestoreFailed
	}
	byBinding := make(map[string]RestoreBindingInput, len(inputs))
	for _, input := range inputs {
		if _, duplicate := byBinding[input.BindingID]; duplicate {
			target.Failed = true
			return ErrRestoreFailed
		}
		byBinding[input.BindingID] = input
	}
	for _, binding := range plan.Bindings {
		input, present := byBinding[binding.BindingID]
		if !present {
			target.Failed = true
			return ErrRestoreFailed
		}
		codec, present := plan.Codecs[input.CodecID]
		if !present || codec.BindingID != binding.BindingID || codec.SHA256 != input.CodecSHA256 || binding.BackupCodecID != codec.CodecID || binding.BackupCodecSHA256 != codec.SHA256 {
			target.Failed = true
			return ErrBackupCodecUnsupported
		}
		if len(input.Entries) > codec.MaxItems {
			target.Failed = true
			return ErrRestoreFailed
		}
		var bindingBytes int64
		for _, entry := range input.Entries {
			if int64(len(entry)) > codec.MaxEntryBytes {
				target.Failed = true
				return ErrRestoreFailed
			}
			bindingBytes += int64(len(entry))
			if bindingBytes > codec.MaxBindingBytes {
				target.Failed = true
				return ErrRestoreFailed
			}
		}
		if err := port.RestoreBinding(ctx, binding, codec, input); err != nil {
			target.Failed = true
			return fmt.Errorf("%w: %s", ErrRestoreFailed, binding.BindingID)
		}
		if err := port.ValidateBinding(ctx, binding); err != nil {
			target.Failed = true
			return fmt.Errorf("%w: %s", ErrRestoreFailed, binding.BindingID)
		}
		delete(byBinding, binding.BindingID)
	}
	if len(byBinding) != 0 {
		target.Failed = true
		return ErrRestoreFailed
	}
	return nil
}

func (target RestoreTarget) MayServe() bool {
	return !target.Failed
}
