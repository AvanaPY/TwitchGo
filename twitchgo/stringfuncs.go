package twitchlib

func LimitStringLength(s string, length int, fill string) string {
    if len(s) > length {
        s = s[:length-len(fill)] + fill
    }
    return s
}
