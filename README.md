## Kscope(Kubescope)
An open system to monitor and test your apis running inside and outside of a kubernetes cluster.

### Introduction
With kubernets being the de facto for container orchestration system. More or more services are enclosed within the kubernetes system. Only the front facing api are exposed and 
most the other apis are accessed within the internal systems. And it makes total sense to access those services directly using the kubernetes service (eg. `svc-name.svc.cluster.local`) instead of going via the internet.

Runscope and assertible.com does a nice job moniting and cheking the journeys of your apis. But how do we check the services residing completely within the kubernetes scope. Well Kubescope is created to handle that. 
With more and more apis running withing kubernets (only handful of services are edge services) testing and monitoring these apis is important. 


