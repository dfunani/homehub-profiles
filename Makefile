# Refresh atlas.sum after adding or editing files under migrations/
.PHONY: atlas-hash
atlas-hash:
	atlas migrate hash --dir "file://$(CURDIR)/migrations"
