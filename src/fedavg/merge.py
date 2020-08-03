import torch

def merge(models, merged_output_path):
    weights = [torch.load(m['path'], 'cpu') for m in models]
    total_data_size = sum(m['size'] for m in models)
    factors = [m['size'] / total_data_size for m in models]

    merged = {}
    for key in weights[0].keys():
        merged[key] = sum([w[key] * f for w, f in zip(weights, factors)])

    torch.save(merged, merged_output_path)
