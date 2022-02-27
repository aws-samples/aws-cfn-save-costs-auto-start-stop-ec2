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
            'Name': 'tag:AutoStart',
            'Values': ['TRUE','True','true']
        },
        {
            'Name': 'instance-state-name',
            'Values': ['stopped']
        }
    ]

    instances = ec2.instances.filter(Filters=filters)
    StoppedInstances = [instance.id for instance in instances]
    print("Stopped Instances with AutoStart Tag : " + str(StoppedInstances))

    if len(StoppedInstances) > 0:
        for instance in instances:
            if instance.state['Name'] == 'stopped':
                print("Starting Instance : " + instance.id)
        AutoStarting = ec2.instances.filter(InstanceIds=StoppedInstances).start()
        print("Started Instances : " + str(StoppedInstances))
    else:
        print("Instance not in Stopped state or AutoStart Tag not set...")
