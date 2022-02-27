import boto3
import logging
import os

logger = logging.getLogger()
logger.setLevel(logging.INFO)

region = os.environ['AWS_REGION']
ec2 = boto3.resource('ec2', region_name=region)

def lambda_handler(event, context):

    filters = [
        {
            'Name': 'tag:AutoStop',
            'Values': ['TRUE','True','true']
        },
        {
            'Name': 'instance-state-name',
            'Values': ['running']
        }
    ]

    instances = ec2.instances.filter(Filters=filters)
    RunningInstances = [instance.id for instance in instances]
    print("Running Instances with AutoStop Tag : " + str(RunningInstances))

    if len(RunningInstances) > 0:
        for instance in instances:
            if instance.state['Name'] == 'running':
                print("Stopping Instance : " + instance.id)
        AutoStopping = ec2.instances.filter(InstanceIds=RunningInstances).stop()
        print("Stopped Instances : " + str(RunningInstances))
    else:
        print("Instance not in Running state or AutoStop Tag not set...")
