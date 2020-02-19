D=importdata('/home/geforce/Downloads/robocup-data/logs/2019/div-a/all.csv');

timestamp = D(:,1);
team = D(:,2);
robotId = D(:,3);
age = D(:,4);
numFrames = D(:,5);
duration = D(:,6);

bins = (0.1:0.001:0.5);

figure
histogram(duration/1e9, bins)
xlabel('duration [s]')
ylabel('count')
title('Histogram over data loss duration')