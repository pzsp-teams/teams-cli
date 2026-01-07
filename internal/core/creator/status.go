package creator

// Status represents the state of a creation operation.
type Status string

const (
	// StatusCreated indicates the resource was successfully created.
	StatusCreated Status = "Created"

	// StatusWouldCreate indicates the resource would be created in a real run (dry-run mode).
	StatusWouldCreate Status = "WouldCreate"

	// StatusAlreadyExists indicates the resource already exists and no action was taken.
	StatusAlreadyExists Status = "AlreadyExists"

	// StatusFailed indicates the operation failed with an error.
	StatusFailed Status = "Failed"

	// StatusMembersEnsured indicates members were successfully added to an existing resource.
	StatusMembersEnsured Status = "MembersEnsured"

	// StatusWouldEnsureMembers indicates members would be added in a real run (dry-run mode).
	StatusWouldEnsureMembers Status = "WouldEnsureMembers"

	// StatusPartiallyEnsured indicates some but not all members were successfully added.
	StatusPartiallyEnsured Status = "PartiallyEnsured"
)
