# Goal rules for workspace admission (TH-HARNESS-REQ-506). Make parses once:
# each managed recipe takes shared admission itself, through RUN_ADMITTED,
# before any preparation or write; cleanup takes exclusive admission.
ifneq ($(filter command environment override,$(origin MAKECMDGOALS)),)
$(error MAKECMDGOALS cannot override worktree admission)
endif
ifneq ($(filter clean distclean,$(MAKECMDGOALS)),)
ifneq ($(filter-out clean distclean,$(MAKECMDGOALS)),)
$(error cleanup cannot be combined with managed work in one Make invocation)
endif
endif
ifneq ($(filter test-run-status,$(MAKECMDGOALS)),)
ifneq ($(words $(MAKECMDGOALS)),1)
$(error test-run-status must be the sole Make goal)
endif
endif
