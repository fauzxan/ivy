package central

/*
In the event of a failure, this message is sent as a last resort to the backup
manager.
*/
type SelfCopyMessage struct {
	Central Central
}
