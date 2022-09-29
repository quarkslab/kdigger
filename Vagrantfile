# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "generic/ubuntu2004"
  config.vm.hostname = "kdigger"
  config.vm.define "kdigger"

  config.vm.network "private_network", type: "dhcp"

  config.vm.synced_folder "./", "/home/vagrant/kdigger"

  config.vm.provider "virtualbox" do |vb|
    vb.cpus = 6
    vb.memory = "4096"
  end

  # Install docker
  config.vm.provision :docker

  config.vm.provision "shell", inline: <<-SHELL
    apt-get update
    apt-get install -y build-essential curl neovim git

    GO_VERSION=1.19.1
    echo "Install Go $GO_VERSION"
    curl -OL https://golang.org/dl/go$GO_VERSION.linux-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go$GO_VERSION.linux-amd64.tar.gz
    rm -f go$GO_VERSION.linux-amd64.tar.gz
    echo 'PATH=$PATH:/usr/local/go/bin' >> /home/vagrant/.bashrc
    echo 'PATH=$PATH:/home/vagrant/go/bin' >> /home/vagrant/.bashrc

    echo "Install arkade"
    curl -sLS https://get.arkade.dev | sudo sh
    echo 'PATH=$PATH:$HOME/.arkade/bin/' >> /home/vagrant/.bashrc

  SHELL

  config.vm.provision "shell", privileged: false, inline: <<-SHELL
    echo "Install golangci-lint"
    GOLANGCI_LINT_VERSION=1.49.0
    export PATH=$PATH:/usr/local/go/bin
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v$GOLANGCI_LINT_VERSION
  SHELL
end

