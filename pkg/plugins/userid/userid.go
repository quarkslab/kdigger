package userid

import (
	"os/user"

	"github.com/mtardy/kdigger/pkg/bucket"
)

const (
	bucketName        = "userid"
	bucketDescription = "UserID retrieves UID, GID and their corresponding names."
)

var bucketAliases = []string{"userids", "id"}

type UserIDBucket struct{}

func (n UserIDBucket) Run() (bucket.Results, error) {
	usr, err := user.Current()
	if err != nil {
		return bucket.Results{}, err
	}

	mainGroup, err := user.LookupGroupId(usr.Gid)
	if err != nil {
		return bucket.Results{}, err
	}

	// the following will always panic without CGO enabled
	// "panic: user: GroupIds requires cgo"

	// otherGroups, err := usr.GroupIds()
	// if err != nil {
	//     return bucket.Results{}, err
	// }

	res := bucket.NewResults(bucketName)
	res.SetHeaders([]string{"userID", "userName", "groupID", "groupName", "homeDir"})
	res.AddContent([]interface{}{usr.Uid, usr.Username, usr.Gid, mainGroup.Name, usr.HomeDir})

	return *res, nil
}

// Register registers a plugin
func Register(b *bucket.Buckets) {
	b.Register(bucketName, bucketAliases, bucketDescription, func(config bucket.Config) (bucket.Interface, error) {
		return NewUserIDBucket(config)
	})
}

func NewUserIDBucket(config bucket.Config) (*UserIDBucket, error) {
	return &UserIDBucket{}, nil
}
