package event

// Int method is used to grab the int32 enum value of the Level in the Event
//
// This value may be used for log-level-filtering
func (l Level) Int() int32 {
	return Level_value[l.String()]
}
