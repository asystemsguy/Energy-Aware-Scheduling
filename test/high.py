import numpy as np
import pandas as pd
from scipy.cluster.hierarchy import dendrogram, linkage  
from matplotlib import pyplot as plt
import json
import codecs

# Make all the not connected nodes as 1 by default
# ToDo read these values from a JSON

Dist = np.array([
    [1,9,6,1,1],  
    [9,1,1,1,5],
    [6,1,1,1,1],
    [1,1,1,1,8],
    [1,5,1,8,1],
 ])

services_bw = codecs.open("services_bw.json", 'r', encoding='utf-8').read()
services_bw_json = json.loads(services_bw)
Dist = np.array(services_bw_json["services"])
num_services = len(Dist)

# Just takes the index from the array and returns back the distance from distance array
def calculate_distance(u, v, w=None):
    return (1/Dist[int(u[0])][int(v[0])])

def display_linkage_matrix(linked):
    # Display dendrogram
    labelList = range(0, num_services)

    plt.figure(figsize=(10, 7))  
    dendrogram(linked,  
                orientation='top',
                labels=labelList,
                distance_sort='descending',
                show_leaf_counts=True)
    plt.show() 


def extract_levels(row_clusters, labels):
    clusters = {}
    for row in range(row_clusters.shape[0]):
        cluster_n = row + len(labels)
        # which clusters / labels are present in this row
        glob1, glob2 = row_clusters[row, 0], row_clusters[row, 1]

        # if this is a cluster, pull the cluster
        this_clust = []
        for glob in [glob1, glob2]:
            if glob > (len(labels)-1):
                this_clust += clusters[glob]
            # if it isn't, add the label to this cluster
            else:
                this_clust.append(glob)

        clusters[cluster_n] = this_clust
    return clusters

# Generate this array based on number of services to index distance array
# For 5 Services it will look 
# np.array([
#   [0,0],
#   [1,0],
#   [2,0],
#   [3,0],
#   [4,0],
#   ])

X = np.zeros((num_services,2))
X[:,0] = np.arange(num_services)

# Single is distance between two nearest neighbours will be concedered fro merge
linked = linkage(X, 'single', metric=calculate_distance)

#Create dict
json_file = {}
json_file['service_clusters'] = extract_levels(linked,X)

#print(json_file)

with open("service_clusters.json", "w") as write_file:   
    #Dump data dict to jason
    write_file.write(json.dumps(json_file))

print(len(linked))





