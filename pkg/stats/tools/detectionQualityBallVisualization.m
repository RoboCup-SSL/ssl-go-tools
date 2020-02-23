year='2019';
division='div-a';
D=importdata(strcat('logs/', year, '/', division, '/ball.csv'));

timestamp = D(:,1);
age = D(:,2);
numFrames = D(:,3);
duration = D(:,4);

bins = (0.05:0.001:0.5);

figure
histogram(duration/1e9, bins)
xlabel('duration [s]')
ylabel('count')
title(strcat(year, ' Histogram over robot data loss duration'))