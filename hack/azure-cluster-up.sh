while getopts "n:r:l:" flag;do
    case "${flag}" in
        n) 
            CLUSTER_NAME=$OPTARG
            echo "Cluster Name: $CLUSTER_NAME"
            ;;
        r) 
            RESOURCE_GROUP=$OPTARG
            echo "Resource Group: $RESOURCE_GROUP"
            ;;
        l) 
            LOCATION=$OPTARG
            echo "Location: $LOCATION"
            ;;
    esac
done

echo $PWD

#check mandatory flags
if [ -z $CLUSTER_NAME ]; then
    echo "Error: Missing argument Cluster Name (flag -n)"
    exit 1
fi
if [ -z $RESOURCE_GROUP ]; then
    echo "Error: Missing argument Resource Group Name (flag -r)"
    exit 1
fi

echo -e "\nChecking if Resource Group $RESOURCE_GROUP Exists"
if [ $(az group exists -n $RESOURCE_GROUP) = true ];
then
    echo "$RESOURCE_GROUP exists"
else
    echo "$RESOURCE_GROUP does not exist"
    echo "Creating new Resource Group $RESOURCE_GROUP"
    if [ -z $LOCATION]; then
        echo "Error: Cannot create Resource group without Location (flag -l)"
        exit 1
    fi
    az group create -l $LOCATION -n $RESOURCE_GROUP
fi

echo -e "\nSpinning up Azure Kubernetes Cluster $CLUSTER_NAME"
az aks create --resource-group $RESOURCE_GROUP --name $CLUSTER_NAME --enable-addons monitoring --generate-ssh-keys

echo -e "\nGetting Credentials for Cluster $CLUSTER_NAME"
az aks get-credentials --resource-group $RESOURCE_GROUP --name $CLUSTER_NAME
echo -e "\n"

DRIVER_NAME=$(dirname "$(realpath ${BASH_SOURCE[0]})")
source "$DRIVER_NAME/cosi-install.sh"