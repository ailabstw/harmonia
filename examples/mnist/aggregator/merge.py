from net import Net
import torch
from torchvision import datasets, transforms
import torch.nn.functional as F
import logging

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


def merge(inputs, output, test_batch_size=1000, no_cuda=False):
    """
    inputs: filenames of checkpoints to merge
    output: output checkpoint filename
    test_batch_size: input batch size for testing (default: 1000)
    no_cuda: disables CUDA training
    """
    model = Net()

    logging.info("inputs: {}".format(inputs))
    logging.info("output: {}".format(output))

    # TODO: hard-coded for MNIST
    sum_conv1 = []
    sum_conv2 = []
    sum_fc1 = []
    sum_fc2 = []
    use_cuda = not no_cuda and torch.cuda.is_available()
    for f in inputs:
        if not use_cuda:
            data = torch.load(f, map_location='cpu')
        else:
            data = torch.load(f)
        m = Net()
        m.load_state_dict(data['model_state_dict'])

        sum_conv1.append(m.conv1.weight)
        sum_conv2.append(m.conv2.weight)
        sum_fc1.append(m.fc1.weight)
        sum_fc2.append(m.fc2.weight)

    model.conv1.weight = torch.nn.parameter.Parameter(
        sum(sum_conv1) / len(inputs))
    model.conv2.weight = torch.nn.parameter.Parameter(
        sum(sum_conv2) / len(inputs))
    model.fc1.weight = torch.nn.parameter.Parameter(
        sum(sum_fc1) / len(inputs))
    model.fc2.weight = torch.nn.parameter.Parameter(
        sum(sum_fc2) / len(inputs))

    device = torch.device("cuda" if use_cuda else "cpu")
    kwargs = {'num_workers': 1, 'pin_memory': True} if use_cuda else {}
    test_loader = torch.utils.data.DataLoader(datasets.MNIST(
        '../data',
        train=False,
        download=True,
        transform=transforms.Compose([
            transforms.ToTensor(),
            transforms.Normalize((0.1307, ), (0.3081, ))
        ])),
                                              batch_size=test_batch_size,
                                              shuffle=True,
                                              **kwargs)
    __test(model, device, test_loader)

    torch.save({
        'model_state_dict': model.state_dict(),
    }, output)
