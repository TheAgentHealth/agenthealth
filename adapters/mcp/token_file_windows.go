package mcp

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

func protectTokenFile(file *os.File) error {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}
	sd, err := windows.SecurityDescriptorFromString("D:P(A;;FA;;;" + user.User.Sid.String() + ")")
	if err != nil {
		return err
	}
	acl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	// Apply a protected owner-only DACL before writing any secret bytes. Setting
	// the owner also handles elevated processes whose default owner is a group.
	return windows.SetNamedSecurityInfo(file.Name(), windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		user.User.Sid, nil, acl, nil)
}
func privateTokenFile(file *os.File) bool {
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return false
	}
	// Validate the open file handle, not a name that can be swapped after open.
	sd, err := windows.GetSecurityInfo(windows.Handle(file.Fd()), windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return false
	}
	owner, _, err := sd.Owner()
	if err != nil || owner == nil || !owner.Equals(user.User.Sid) {
		return false
	}
	control, _, err := sd.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		return false
	}
	acl, _, err := sd.DACL()
	if err != nil || acl == nil || acl.AceCount != 1 {
		return false
	}
	var ace *windows.ACCESS_ALLOWED_ACE
	if windows.GetAce(acl, 0, &ace) != nil || ace == nil || ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || ace.Header.AceFlags != 0 {
		return false
	}
	sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
	return sid.IsValid() && sid.Equals(user.User.Sid)
}
