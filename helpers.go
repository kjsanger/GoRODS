/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"strconv"
	"time"
	"unsafe"
)

func cTimeToTime(cTime *C.char) time.Time {
	unixStamp, _ := strconv.ParseInt(C.GoString(cTime), 10, 64)
	return time.Unix(unixStamp, 0)
}

func timeStringToTime(ts string) time.Time {
	unixStamp, _ := strconv.ParseInt(ts, 10, 64)
	return time.Unix(unixStamp, 0)
}

func getTypeString(typ int) string {
	switch typ {
	case DataObjType:
		return "d"
	case CollectionType:
		return "C"
	case ZoneType:
		return "Z"
	case ResourceType:
		return "R"
	case UserType:
		return "u"
	case AdminType:
		return "a"
	case GroupAdminType:
		return "ga"
	case GroupType:
		return "g"
	case UnknownType:
		return "?"
	case Null:
		return "null"
	case Read:
		return "read"
	case Write:
		return "write"
	case Own:
		return "own"
	case Local:
		return "local"
	case Remote:
		return "remote"
	case Cache:
		return "cache"
	case Archive:
		return "archive"
	default:
		panic(newError(Fatal, "unrecognized type constant"))
	}
}

func aclSliceToResponse(result *C.goRodsACLResult_t, con *Connection) (ACLs, error) {
	defer C.gorods_free_acl_result(result)

	unsafeArr := unsafe.Pointer(result.aclArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.goRodsACL_t
	slice := (*[1 << 30]C.goRodsACL_t)(unsafeArr)[:arrLen:arrLen]

	response := make(ACLs, 0)

	for _, acl := range slice {

		typeString := C.GoString(acl.acltype)

		typeMap := map[string]int{
			"rodsgroup":  GroupType,
			"rodsuser":   UserType,
			"rodsadmin":  AdminType,
			"groupadmin": GroupAdminType,
		}
		var aclType int = typeMap[typeString]
		if aclType == 0 {
			aclType = UnknownType
		}

		accessString := C.GoString(acl.dataAccess)
		var accessLevel int
		switch accessString {
		case "own":
			accessLevel = Own
		case "modify object":
			accessLevel = Write
		case "read object":
			accessLevel = Read
		default:
			accessLevel = Null
		}

		var accessObject AccessObject
		if aclType == UserType || aclType == AdminType || aclType == GroupAdminType {
			if usrs, err := con.GetUsers(); err == nil {
				if existingUsr := usrs.FindByName(C.GoString(acl.name)); existingUsr != nil {
					accessObject = existingUsr
				} else {
					return nil, newError(Fatal, fmt.Sprintf("iRods GetACL Failed: can't find iRODS user by string"))
				}
			} else {
				return nil, err
			}
		} else if aclType == GroupType {
			if grps, err := con.GetGroups(); err == nil {
				if existingGrp := grps.FindByName(C.GoString(acl.name)); existingGrp != nil {
					accessObject = existingGrp
				} else {
					return nil, newError(Fatal, fmt.Sprintf("iRods GetACL Failed: can't find iRODS group by string"))
				}
			} else {
				return nil, err
			}
		} else if aclType == UnknownType {
			return nil, newError(Fatal, fmt.Sprintf("iRods GetACL Failed: Unknown Type"))
		}

		response = append(response, &ACL{
			AccessObject: accessObject,
			AccessLevel:  accessLevel,
			Type:         aclType,
		})

	}

	return response, nil
}

func isString(obj interface{}) bool {
	switch obj.(type) {
	case string:
		return true
	default:
	}

	return false
}
