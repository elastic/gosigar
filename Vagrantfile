# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "centos/7"
  config.vm.hostname = "centos7"
  config.vm.provision "shell", inline: "mkdir -p /home/vagrant/go"
  config.vm.synced_folder ".", "/home/vagrant/go/src/github.com/cloudfoundry/gosigar",
    type: "nfs",
    nfs_version: 4,
    nfs_udp: false

  config.vm.provision "shell", inline: "chown -R vagrant:vagrant /home/vagrant/go"
  install_go = <<-BASH
  set -e

if [ ! -d "/usr/local/go" ]; then
	cd /tmp && wget https://golang.org/dl/go1.17.3.linux-amd64.tar.gz
	cd /usr/local
	tar xvzf /tmp/go1.17.3.linux-amd64.tar.gz
	echo 'export GOPATH=/home/vagrant/go; export PATH=/usr/local/go/bin:$PATH:$GOPATH/bin' >> /home/vagrant/.bashrc
fi
export GOPATH=/home/vagrant/go
export PATH=/usr/local/go/bin:$PATH:$GOPATH/bin
/usr/local/go/bin/go get -u github.com/onsi/ginkgo/ginkgo
/usr/local/go/bin/go get -u github.com/onsi/gomega;
BASH

  setup_audit = <<-BASH
cat > /etc/audit/rules.d/audit.rules << EOF
-a always,exit -F arch=b32 -S creat -S open -S openat -S truncate -S ftruncate -F exit=-EACCES -F auid>=500 -F auid!=4294967295 -k access
-a always,exit -F arch=b32 -S creat -S open -S openat -S truncate -S ftruncate -F exit=-EPERM -F auid>=500 -F auid!=4294967295 -k access
-a always,exit -F arch=b64 -S creat -S open -S openat -S truncate -S ftruncate -F exit=-EACCES -F auid>=500 -F auid!=4294967295 -k access
-a always,exit -F arch=b64 -S creat -S open -S openat -S truncate -S ftruncate -F exit=-EPERM -F auid>=500 -F auid!=4294967295 -k access
EOF

augenrules
/sbin/service auditd restart
BASH

  config.vm.provision "shell", inline: 'yum install -y git-core wget'
  config.vm.provision "shell", inline: install_go
  config.vm.provision "shell", inline: setup_audit
end