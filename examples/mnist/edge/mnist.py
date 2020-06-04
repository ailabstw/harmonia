from __future__ import print_function
import torch
import torch.nn as nn
import torch.nn.functional as F
import torch.optim as optim
from torchvision import datasets, transforms
from torch.optim.lr_scheduler import StepLR
import logging

from net import Net


def __train(model, device, train_loader, optimizer, epoch, log_interval):
    model.train()
    for batch_idx, (data, target) in enumerate(train_loader):
        data, target = data.to(device), target.to(device)
        optimizer.zero_grad()
        output = model(data)
        loss = F.nll_loss(output, target)
        loss.backward()
        optimizer.step()
        if batch_idx % log_interval == 0:
            logging.info('Train Epoch: {} [{}/{} ({:.0f}%)]\tLoss: {:.6f}'.format(
                epoch, batch_idx * len(data), len(train_loader.dataset),
                100. * batch_idx / len(train_loader), loss.item()))


def __test(model, device, test_loader):
    model.eval()
    test_loss = 0
    correct = 0
    with torch.no_grad():
        for data, target in test_loader:
            data, target = data.to(device), target.to(device)
            output = model(data)
            test_loss += F.nll_loss(
                output, target, reduction='sum').item()  # sum up batch loss
            pred = output.argmax(
                dim=1,
                keepdim=True)  # get the index of the max log-probability
            correct += pred.eq(target.view_as(pred)).sum().item()

    test_loss /= len(test_loader.dataset)

    logging.info(
        '\nTest set: Average loss: {:.4f}, Accuracy: {}/{} ({:.0f}%)\n'.format(
            test_loss, correct, len(test_loader.dataset),
            100. * correct / len(test_loader.dataset)))

def train(data_slice: list, output: str, batch_size=64, test_batch_size=1000,
          epochs=1, lr=1.0, gamma=0.7, no_cuda=False, seed=1,
          log_interval=10, resume=''):
    """
    PyTorch MNIST Example
    data_slice: index of MNIST dat for training, should be in range [0:60000)
    output: output checkpoint filename
    batch_size: input batch size for training (default: 64)
    test_batch_size: input batch size for testing (default: 1000)
    epochs: number of epochs to train (default: 10)
    lr: learning rate (default: 1.0)
    gamma: Learning rate step gamma (default: 0.7)')
    no_cuda: disables CUDA training
    seed: random seed (default: 1)
    log_interval: how many batches to wait before logging training status
    resume: filename of resume from checkpoint
    """
    use_cuda = not no_cuda and torch.cuda.is_available()

    torch.manual_seed(seed)

    device = torch.device("cuda" if use_cuda else "cpu")

    logging.info("[MNIST] Training data loading...")
    kwargs = {'num_workers': 1, 'pin_memory': True} if use_cuda else {}
    training_data = datasets.MNIST('../data',
                                   train=True,
                                   download=True,
                                   transform=transforms.Compose([
                                       transforms.ToTensor(),
                                       transforms.Normalize((0.1307, ),
                                                            (0.3081, ))
                                   ]))
    train_loader = torch.utils.data.DataLoader(torch.utils.data.Subset(
        training_data, list(range(2500))),
                                               batch_size=batch_size,
                                               shuffle=True,
                                               **kwargs)
    test_loader = torch.utils.data.DataLoader(datasets.MNIST(
        '../data',
        train=False,
        transform=transforms.Compose([
            transforms.ToTensor(),
            transforms.Normalize((0.1307, ), (0.3081, ))
        ])),
                                              batch_size=test_batch_size,
                                              shuffle=True,
                                              **kwargs)

    model = Net().to(device)
    optimizer = optim.Adadelta(model.parameters(), lr=lr)

    try:
        resume = torch.load(resume)
        model.load_state_dict(resume['model_state_dict'])
        if 'optimizaer_state_dict' in resume:
            optimizer.load_state_dict(resume['optimizer_state_dict'])
        if 'epoch' in resume:
            epoch = resume['epoch']
        train_loader = torch.utils.data.DataLoader(torch.utils.data.Subset(
            training_data, data_slice),
            batch_size=batch_size,
            shuffle=True,
            **kwargs)
    except Exception:
        logging.info("[MNIST] Empty Base Model")

    logging.info("[MNIST] Training...")
    model.train()

    scheduler = StepLR(optimizer, step_size=1, gamma=gamma)
    for epoch in range(1, epochs + 1):
        __train(model, device, train_loader, optimizer, epoch, log_interval)
        __test(model, device, test_loader)
        scheduler.step()

    logging.info("[MNIST] Save Weights... [{}]".format(output))
    torch.save(
        {
            'epoch': epoch,
            'model_state_dict': model.state_dict(),
            'optimizer_state_dict': optimizer.state_dict(),
        }, output)
