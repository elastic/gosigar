//go:build linux
// +build linux

package gosigar_test

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	sigar "github.com/elastic/gosigar"
	"github.com/stretchr/testify/assert"
)

func setUpCommonTest(t testing.TB) {
	oldMtabf := sigar.Mtabf
	sigar.Mtabf = filepath.Join(t.TempDir(), "mtab")
	t.Cleanup(func() { sigar.Mtabf = oldMtabf })
}

func TestLinuxFileSystemList(t *testing.T) {
	setUpCommonTest(t)

	var mtabLines = []string{
		"sysfs /sys sysfs rw,nosuid,nodev,noexec,relatime 0 0",
		"proc /proc proc rw,nosuid,nodev,noexec,relatime 0 0",
		"udev /dev devtmpfs rw,nosuid,relatime,size=32441540k,nr_inodes=8110385,mode=755,inode64 0 0",
		"devpts /dev/pts devpts rw,nosuid,noexec,relatime,gid=5,mode=620,ptmxmode=000 0 0",
		"tmpfs /run tmpfs rw,nosuid,nodev,noexec,relatime,size=6496012k,mode=755,inode64 0 0",
		"/dev/mapper/ubuntu--vg-root / ext4 rw,relatime,errors=remount-ro 0 0",
		"securityfs /sys/kernel/security securityfs rw,nosuid,nodev,noexec,relatime 0 0",
		"cgroup2 /sys/fs/cgroup cgroup2 rw,nosuid,nodev,noexec,relatime,nsdelegate,memory_recursiveprot 0 0",
		"pstore /sys/fs/pstore pstore rw,nosuid,nodev,noexec,relatime 0 0",
		"efivarfs /sys/firmware/efi/efivars efivarfs rw,nosuid,nodev,noexec,relatime 0 0",
		"bpf /sys/fs/bpf bpf rw,nosuid,nodev,noexec,relatime,mode=700 0 0",
		"systemd-1 /proc/sys/fs/binfmt_misc autofs rw,relatime,fd=32,pgrp=1,timeout=0,minproto=5,maxproto=5,direct,pipe_ino=2121 0 0",
		"debugfs /sys/kernel/debug debugfs rw,nosuid,nodev,noexec,relatime 0 0",
		"mqueue /dev/mqueue mqueue rw,nosuid,nodev,noexec,relatime 0 0",
		"hugetlbfs /dev/hugepages hugetlbfs rw,nosuid,nodev,relatime,pagesize=2M 0 0",
		"tracefs /sys/kernel/tracing tracefs rw,nosuid,nodev,noexec,relatime 0 0",
		"configfs /sys/kernel/config configfs rw,nosuid,nodev,noexec,relatime 0 0",
		"fusectl /sys/fs/fuse/connections fusectl rw,nosuid,nodev,noexec,relatime 0 0",
		"/dev/loop0 /snap/bare/5 squashfs ro,nodev,relatime,errors=continue,threads=single 0 0",
		"/dev/nvme0n1p2 /boot ext2 rw,relatime,stripe=4 0 0",
		"/dev/nvme0n1p1 /boot/efi vfat rw,relatime,fmask=0077,dmask=0077,codepage=437,iocharset=iso8859-1,shortname=mixed,errors=remount-ro 0 0",
		"binfmt_misc /proc/sys/fs/binfmt_misc binfmt_misc rw,nosuid,nodev,noexec,relatime 0 0",
		"nsfs /run/snapd/ns/cups.mnt nsfs rw 0 0",
		"overlay /var/lib/docker/overlay2/7cc388859979000930e89f3415bb81782f72b112b7e41afe69bcb6c13166f2ae/merged overlay rw,relatime,lowerdir=/var/lib/docker/overlay2/l/AZ3FOIHGPLV275MZ5YV2STWAVR:/var/lib/docker/overlay2/l/WGXYCW3I3L4PPXKPGF7RN5JGYG:/var/lib/docker/overlay2/l/6I2NXU5RUZN4FS5K7GEMA5EGTA:/var/lib/docker/overlay2/l/7DOY3UP2A36LPIWWW3ELAXVVLM:/var/lib/docker/overlay2/l/6ZQG7QAA3R5DOT27IN7POBA2VW:/var/lib/docker/overlay2/l/2HA2K65HUPOXP3I7U6OSDEF7DU:/var/lib/docker/overlay2/l/ILPOWEXY3SJKRJOFBDMKS3NPDQ:/var/lib/docker/overlay2/l/LHTHTFA7B5P7A4V2CE7ZFETMFW,upperdir=/var/lib/docker/overlay2/7cc388859979000930e89f3415bb81782f72b112b7e41afe69bcb6c13166f2ae/diff,workdir=/var/lib/docker/overlay2/7cc388859979000930e89f3415bb81782f72b112b7e41afe69bcb6c13166f2ae/work,nouserxattr 0 0",
		"overlay /var/lib/docker/overlay2/229ea374b94e0a60368c0759cc104aa3ef44e7bd46e8e389603307673aab34e7/merged overlay rw,relatime,lowerdir=/var/lib/docker/overlay2/l/25NOBV3WHXG5P3GRO2FRUCWAB3:/var/lib/docker/overlay2/l/NORNMK2W42ELHAA4GJWQX5ZOAW:/var/lib/docker/overlay2/l/RDVR4KEE5JQPI5SDRG333UL7RR:/var/lib/docker/overlay2/l/NDA6EGGA5TVHNZYRLSNACMCQWF:/var/lib/docker/overlay2/l/PSHEG744SP6T5QTAYIBKOHQRTF:/var/lib/docker/overlay2/l/CVJ6EWWFZZB4XNNCTTX2OWOKC7:/var/lib/docker/overlay2/l/OYEE2F2W4IA5RFZE6ZBTHJIUSH:/var/lib/docker/overlay2/l/ZRFGD5OZAXDXEACSMEYWGGSBX7:/var/lib/docker/overlay2/l/CILYIYAPTFNWP7UM7S55KQSI4D:/var/lib/docker/overlay2/l/F4PHFL4L2NAF5GJSJIGQPERKPU:/var/lib/docker/overlay2/l/5R62UHZSGCI2AH7SYFUHVMH56E:/var/lib/docker/overlay2/l/2SENKE2UI73NBAGTGIVH6K3O2A:/var/lib/docker/overlay2/l/DWRXBEGY7U3IIIBVEA6V2AKRBJ:/var/lib/docker/overlay2/l/3GUSTJRGFDDYXLT7V44I6YB3XN:/var/lib/docker/overlay2/l/DEC4B5I67LMALKVPQJAFP5HX46:/var/lib/docker/overlay2/l/ZI5YOAKJUTPBKKOMMTWKITYIXS:/var/lib/docker/overlay2/l/RKWKKH46TJP7S6ZERYR5SPS73N:/var/lib/docker/overlay2/l/LQTPCWBTTXHTHKNCQZF2E2D6AP:/var/lib/docker/overlay2/l/HUNN3R34OVJSNKILBJSMTJTIFO:/var/lib/docker/overlay2/l/S6DGOW3IPS2PT463PG7SYRMUJX:/var/lib/docker/overlay2/l/GKZU26J5SSSQH2RBIMKYUDUNWQ:/var/lib/docker/overlay2/l/SPS2VYCUOWEKCA47Q2FL3T3ZVO:/var/lib/docker/overlay2/l/DJ7CFAJ64XB7AFGAM5UZ3MEOHT:/var/lib/docker/overlay2/l/MJOQ7IFA5JLZ4APQHPRA2GFIKF:/var/lib/docker/overlay2/l/DVAUELMLR5VFOLXLCG4RHVJJH4:/var/lib/docker/overlay2/l/DDV4OD36D4CACTBTVXQ7CQVJTK:/var/lib/docker/overlay2/l/TQSEOKAR5Q4QJI52HJH6TWSENX:/var/lib/docker/overlay2/l/ZJPLLJWA2EDO2QFO6JL5YU5NXC:/var/lib/docker/overlay2/l/NW7MD4MUYBUC27L6XYJTKELXZU:/var/lib/docker/overlay2/l/HF3FA5NGW27JT43D44L4UFWSUS:/var/lib/docker/overlay2/l/MG27ZBUPW3NNRRM4TPE672DPKZ:/var/lib/docker/overlay2/l/IBM7RSWRDIZJPZNAOOW6H3SQUW:/var/lib/docker/overlay2/l/GJI7YYX4T3YCSS35JXQVOLMUPU:/var/lib/docker/overlay2/l/POIXRSHOGGH5SRY6IKBSO7WM6E:/var/lib/docker/overlay2/l/C3HJJ3QW33WNHC33X5VFLUT3DR:/var/lib/docker/overlay2/l/CZNUABCWBZ6T6Z636WZUKBJY3L:/var/lib/docker/overlay2/l/Z2KSHZNRNQ4HUHLXOA4ZBE2P3E:/var/lib/docker/overlay2/l/SJTVBG2UJSGT23AVCSDVR4PNVL:/var/lib/docker/overlay2/l/M7T7PG36WVHPGA5J5QFK6E3Z3T:/var/lib/docker/overlay2/l/3K52OVGIRBUIUOQ7A6XKE3CI7Q:/var/lib/docker/overlay2/l/NPAO3V2FJSNSOIR5IJYRXSJUN2:/var/lib/docker/overlay2/l/MJEUBSKJ3EIEXKI63O23MVJP6L:/var/lib/docker/overlay2/l/KI4KHMEL5YIOHU6M5HJOOP7SOA:/var/lib/docker/overlay2/l/OJ2AAUKW7A27XRU3A6JLFSXPED:/var/lib/docker/overlay2/l/JMSLHEDCQMFBEPIPZGKKPUL2GZ:/var/lib/docker/overlay2/l/63VRX6JDKMNOBZTXFFXGTWLHH4:/var/lib/docker/overlay2/l/WATKHFOLY7ATZ6OSPN63M4SMTN:/var/lib/docker/overlay2/l/2OGULZVZAUODJXF3ZKMTMIRYWC:/var/lib/docker/overlay2/l/2NZZ4KSCVJKL2GCIHHXWPYY3VC:/var/lib/docker/overlay2/l/XOMQEH3JRC4GCBK4ZKEKTISBZV:/var/lib/docker/overlay2/l/SJHHUSXZQ4ZPEZGOFYJ2B7FZHK:/var/lib/docker/overlay2/l/RFETRSPAIPUYYLBY4UWAUEI4WD:/var/lib/docker/overlay2/l/FBF7BNUTJTB4QXRBS4IMF3XTA5:/var/lib/docker/overlay2/l/FEPH4I477I3WOOLYHEFUDJYNYC:/var/lib/docker/overlay2/l/JNZJLFLK6EBJOXBVRY423B5WCW:/var/lib/docker/overlay2/l/QXPFXSXPE3OFAVKJMKEU5OBLGF:/var/lib/docker/overlay2/l/X4LAMESX5ANCAMDB3H5R7EEAGH:/var/lib/docker/overlay2/l/6EM7RPQGAT2QUCPZTOYQL2YA7W:/var/lib/docker/overlay2/l/4YCKFX7RIZSE3X7NR33L2AMXTF:/var/lib/docker/overlay2/l/J4IAXBV7AJD4UXI6E2DKBI5USH:/var/lib/docker/overlay2/l/U4LYG45JPNMLITFZPXRZ4YORNH:/var/lib/docker/overlay2/l/U7YWJNILQ4S3E2GMNKVBDU3HE2:/var/lib/docker/overlay2/l/4Q7NTRFPH6PISJPCGL6MSNKXAB:/var/lib/docker/overlay2/l/S4Q5LW6NPUGGV4PNH7RGIZULDB:/var/lib/docker/overlay2/l/JH4VWREORYUTBEG5OWJGRUTSWZ:/var/lib/docker/overlay2/l/DOSO5DTS2LYOPZNKNNDA6UXJVN:/var/lib/docker/overlay2/l/GI43U22RIMZQU2WCKLILS7VIWU:/var/lib/docker/overlay2/l/ELJJ5B5BDXE2XLUU3SXA7NCFV6:/var/lib/docker/overlay2/l/WSEEZO7VSPHRUZV6ZFJZFI2JOJ:/var/lib/docker/overlay2/l/Q7CHINR3TSEIMNYBXBQEM7GV4X:/var/lib/docker/overlay2/l/OJQEGWPTEYASKMEKOJ2APQ47XI,upperdir=/var/lib/docker/overlay2/229ea374b94e0a60368c0759cc104aa3ef44e7bd46e8e389603307673aab34e7/diff,workdir=/var/lib/docker/overlay2/229ea374b94e0a60368c0759cc104aa3ef44e7bd46e8e389603307673aab34e7/work 0 0",
		"portal /run/user/1000/doc fuse.portal rw,nosuid,nodev,relatime,user_id=1000,group_id=1000 0 0",
		"gvfsd-fuse /run/user/1000/gvfs fuse.gvfsd-fuse rw,nosuid,nodev,relatime,user_id=1000,group_id=1000 0 0",
		"/dev/loop75 /snap/discord/225 squashfs ro,nodev,relatime,errors=continue,threads=single 0 0",
	}

	expectedList := []sigar.FileSystem{}
	for _, n := range mtabLines {
		func() {
			fields := strings.Fields(n)
			expectedList = append(expectedList, sigar.FileSystem{
				DirName:     fields[1],
				DevName:     fields[0],
				TypeName:    "",
				SysTypeName: fields[2],
				Options:     fields[3],
				Flags:       0,
			})
		}()
	}
	expected := sigar.FileSystemList{
		List: expectedList,
	}
	writeMtab(mtabLines, sigar.Mtabf)
	fslist := sigar.FileSystemList{}
	if assert.NoError(t, fslist.Get()) {
		assert.Equal(t, expected, fslist)
	}
}

func writeMtab(mtabLines []string, path string) error {
	var mtabBuilder strings.Builder
	for _, n := range mtabLines {
		mtabBuilder.WriteString(n)
		mtabBuilder.WriteString("\n")
	}
	mtabContents := []byte(mtabBuilder.String())
	return ioutil.WriteFile(path, mtabContents, 0700)
}
