D=importdata('/tmp/visionTiming.csv');

timestamp = D(:,1);
camId = D(:,2);
tCapture = D(:,3);
tSent = D(:,4);

timestampDt = diff(timestamp) / 1e9;
tCaptureDt = diff(tCapture);
tSentDt = diff(tSent);

figure
subplot(3,1,1)
plot(timestampDt)
subplot(3,1,2)
plot(tCaptureDt)
subplot(3,1,3)
plot(tSentDt)