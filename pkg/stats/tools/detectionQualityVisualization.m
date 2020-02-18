D=importdata('/home/geforce/Downloads/robocup-data/logs/2019/div-a/all.csv');

timestamp = D(:,1);
team = D(:,2);
robotId = D(:,3);
age = D(:,4);
numFrames = D(:,5);
duration = D(:,6);

% duration = rmoutliers(duration, 'quartiles');

figure
histogram(duration/1e6, (0:1:100))
% subplot(3,1,1)
% plot(timestampDt)
% subplot(3,1,2)
% plot(tCaptureDt)
% subplot(3,1,3)
% plot(tSentDt)