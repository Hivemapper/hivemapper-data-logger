# Kalman filtering GPS data and Accelerometer data

k: time <br/>
x&#770;<sub>k-1</sub>: best estimate at k-1 time <br/>
P<sub>k-1</sub>: covariance matrix at k-1 time </br>
F<sub>k</sub>: prediction matrix which takes every point in our original estimate and moves it to a new predicted location, which is where the system would move if that original estimate was the right one <br/>
B<sub>k</sub>: control matrix which represents some changes that aren't related to the state itself - external influence </br>
u&#8407;<sub>k</sub>: control vector which represents additional information about what’s going on in the world </br>
Q<sub>k</sub>: noise covariance matrix</br>
x&#770;<sub>k</sub>: best estimate at k time <br/>
P<sub>k</sub>: covariance matrix at k time </br>
H<sub>k</sub>: noise covariance matrix </br>
R<sub>k</sub>: covariance matrix of uncertainty </br>
z&#8407;<sub>k</sub>: mean equal to the reading we observed </br>
x&#770;'<sub>k</sub>: best estimate produced after update and which will be the k-1 in the next step <br/>
P'<sub>k</sub>: covariance matrix produced after update and which will be the k-1 in the next step </br>

![kalflow.png](./doc/kalflow.png)

Reference documentation: https://www.bzarg.com/p/how-a-kalman-filter-works-in-pictures/