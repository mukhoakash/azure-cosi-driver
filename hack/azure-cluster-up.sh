while getopts "n:r:l:" flag;do
    case "${flag}" in
        n) 
            CLUSTER_NAME=$OPTARG
            echo "Name: $CLUSTER_NAME"
            ;;
        r) 
            RESOURCE_GROUP=$OPTARG
            echo "Resource Group: $RESOURCE_GROUP"
            ;;
    esac
done
echo -e "\n"

echo "Spinning up Azure Kubernetes Cluster $CLUSTER_NAME"
az aks create --resource-group $RESOURCE_GROUP --name $CLUSTER_NAME --enable-addons monitoring --generate-ssh-keys
echo -e "\n"

echo "Getting Credentials for Cluster $CLUSTER_NAME"
az aks get-credentials --resource-group $RESOURCE_GROUP --name $CLUSTER_NAME
echo -e "\n"

echo "Getting CRD's for COSI and COSI Controller"
kubectl create -k github.com/kubernetes-sigs/container-object-storage-interface-api
kubectl create -k github.com/kubernetes-sigs/container-object-storage-interface-controller
echo -e "\n"