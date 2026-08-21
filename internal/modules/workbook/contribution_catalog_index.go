package workbook

type contributionCatalogIndexes struct {
	queries     map[string]QueryProvider
	creates     map[string]CreateProvider
	patches     map[string]PatchProvider
	conflicts   map[string]ConflictProvider
	clipboards  map[string]ClipboardProvider
	bulk        map[string]BulkProvider
	linkedNotes map[string]LinkedNoteProvider
	supersedes  map[string]SupersedeProvider
}

func newWorkbookContributionCatalog(indexes contributionCatalogIndexes) *WorkbookContributionCatalog {
	return &WorkbookContributionCatalog{
		queries:     indexes.queries,
		creates:     indexes.creates,
		patches:     indexes.patches,
		conflicts:   indexes.conflicts,
		clipboards:  indexes.clipboards,
		bulk:        indexes.bulk,
		linkedNotes: indexes.linkedNotes,
		supersedes:  indexes.supersedes,
	}
}
