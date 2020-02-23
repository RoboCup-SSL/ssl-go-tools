year='2019';
division='div-a';
D=importdata(strcat('logs/', year, '/', division, '/robots.csv'));

timestamp = D(:,1);
team = D(:,2);
robotId = D(:,3);
age = D(:,4);
numFrames = D(:,5);
duration = D(:,6);

bins = (0.05:0.001:0.5);

figure
histogram(duration/1e9, bins)
xlabel('duration [s]')
ylabel('count')
title(strcat(year, ' Histogram over robot data loss duration'))


figure
histogram(timestamp)
xlabel('timestamp [ns]')
ylabel('count')
title(strcat(year, ' Histogram over robot data loss timestamp'))