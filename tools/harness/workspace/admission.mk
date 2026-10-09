# Gate the complete Make invocation, including file prerequisites and preparation.
# Only a live inherited kernel lock can select the admitted branch.
override workspace_shell_quote = '$(subst ','"'"',$(1))'
ifneq ($(filter command environment override,$(origin MAKECMDGOALS)),)
$(error MAKECMDGOALS cannot override worktree admission)
endif
override workspace_cleanup_goals := $(filter clean distclean,$(MAKECMDGOALS))
ifneq ($(workspace_cleanup_goals),)
ifneq ($(filter-out clean distclean,$(MAKECMDGOALS)),)
$(error cleanup cannot be combined with managed work in one Make invocation)
endif
override workspace_admitted := yes
else ifneq ($(filter test-run-status,$(MAKECMDGOALS)),)
ifneq ($(words $(MAKECMDGOALS)),1)
$(error test-run-status must be the sole Make goal)
endif
# This read-only observer must neither create a lock nor delay cleanup.
override workspace_admitted := yes
else
override workspace_admitted := $(if $(filter shared,$(shell bash tools/harness/workspace/admission.sh verify 2>/dev/null)),yes,no)
endif
ifeq ($(workspace_admitted),no)
.DEFAULT_GOAL := workspace-admission
.PHONY: workspace-admission $(MAKECMDGOALS)
workspace-admission:
	+@bash tools/harness/workspace/admission.sh run -- $(MAKE) --no-print-directory $(foreach goal,$(MAKECMDGOALS),$(call workspace_shell_quote,$(goal)))
$(MAKECMDGOALS): workspace-admission ; @:
endif
