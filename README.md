## Automate Start and Stop of Amazon EC2 Instances to Save costs

Saving money is a top priority for any AWS user. You can save money by manually shutting down your servers when they’re not in use. However, manually managing and administering multiple servers is quite difficult and time-consuming. This script offers a solution to automate the server stop and start procedure by scheduling with either a fixed time, a flexible time, or both.

You can stop your instances during non-working hours and start them during working hours by scheduling them automatically with minimal configuration. The solution also requires a one-time configuration of Amazon EC2 tags.

You’re only charged for the hours when the services are running. This solution can help cut your operational costs by stopping resources that are not in use and starting resources when capacity is required. You can follow either implementing CloudFormation (CFN) template or Serverless (SAM) template as explained below.

### CloudFormation (CFN) template -
The CloudFormation template cfn_auto_start_stop_ec2/cfn_auto_start_stop_ec2.yaml automatically creates all the AWS resources required for the Amazon EC2 solution to function. Complete the following steps to create your AWS resources via the CloudFormation template:

1.	On the AWS CloudFormation console, choose Create stack.
2.	Choose With new resources (standard).
3.	Choose Template is ready and choose Upload a template file.
4.	Upload the provided .yaml file and choose Next.
5.	For Stack name, enter cfn-auto-start-stop-ec2.
6.	Modify the parameter values that set the default cron schedule as needed. 
7.	For RegionTZ, choose which Region time zone to use. This is the TimeZone of the Region in which your EC2 instances are deployed and you want to set timings convenient to that particular Timezone.
8.	Choose Next and provide tags, if needed.
9.	Choose Next and review the stack details.
10.	Select the acknowledgement check box because this template creates an IAM role and policy.
11.	Choose Create stack. 
12.	Open the stack and navigate to the Resources tab to track the resource creation status. 

To delete all the resources created via this template, choose the stack on the AWS CloudFormation console and choose Delete. Choose Delete stack to confirm the stack deletion.

## Security

See [CONTRIBUTING](CONTRIBUTING.md#security-issue-notifications) for more information.

## License

This library is licensed under the MIT-0 License. See the LICENSE file.

