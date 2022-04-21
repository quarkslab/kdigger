# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "generic/ubuntu2004"
  config.vm.hostname = "kdigger"
  config.vm.define "kdigger"

  config.vm.network "private_network", type: "dhcp"

  config.vm.synced_folder "./", "/home/vagrant/project"

  config.vm.provider "virtualbox" do |vb|
    vb.cpus = 6
    vb.memory = "4096"
  end

  # Install docker
  config.vm.provision :docker

  config.vm.provision "shell", inline: <<-SHELL
    apt-get update
    apt-get install -y build-essential curl neovim zsh git

    GO_VERSION=1.18.1
    echo "Install Go $GO_VERSION"
    curl -OL https://golang.org/dl/go$GO_VERSION.linux-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go$GO_VERSION.linux-amd64.tar.gz
    rm -f go$GO_VERSION.linux-amd64.tar.gz

    echo "Install arkade"
    curl -sLS https://get.arkade.dev | sudo sh

    echo "Get my dotfiles"
    runuser -l vagrant -c 'sh -c "$(curl -fsLS git.io/chezmoi)" -- init --apply mtardy'
    mv /home/vagrant/bin/chezmoi /usr/local/bin
    rm -rf /home/vagrant/bin
    rm -rf /home/vagrant/.vimrc

    # echo 'PATH=$PATH:/usr/local/go/bin' >> /home/vagrant/.zshrc
    echo 'PATH=$PATH:$HOME/.arkade/bin/' >> /home/vagrant/.zshrc

    chsh --shell /bin/zsh vagrant
  SHELL
end

