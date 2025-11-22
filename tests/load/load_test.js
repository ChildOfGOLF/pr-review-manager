import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { SharedArray } from 'k6/data';

const errorRate = new Rate('errors');
const BASE_URL = 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '30s', target: 5 },
    { duration: '3m30s', target: 5 },
    { duration: '30s', target: 8 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'],
    http_req_failed: ['rate<0.001'],
    errors: ['rate<0.001'],
  },
};

const testRunId = Date.now();
const teamNames = [
  `backend_${testRunId}`,
  `frontend_${testRunId}`,
  `devops_${testRunId}`,
  `qa_${testRunId}`,
  `mobile_${testRunId}`
];

export function setup() {
  const userIds = [];
  const prIds = [];
  
  console.log(`Starting test run ${testRunId}, creating fresh test data...`);
  
  for (let i = 0; i < 5; i++) {
    const teamName = teamNames[i];
    const members = [];
    
    for (let j = 0; j < 10; j++) {
      const userId = `user_${testRunId}_${i}_${j}`;
      members.push({
        user_id: userId,
        username: `User ${i}-${j}`,
        is_active: true,
      });
      userIds.push(userId);
    }
    
    const response = http.post(
      `${BASE_URL}/team/add`,
      JSON.stringify({ team_name: teamName, members }),
      { headers: { 'Content-Type': 'application/json' } }
    );
    
    const created = check(response, { 'team created': (r) => r.status === 201 });
    if (!created) {
      console.error(`Failed to create team ${teamName}: ${response.status} - ${response.body}`);
    }
  }
  
  for (let i = 0; i < 10; i++) {
    const prId = `pr_setup_${testRunId}_${i}`;
    const authorId = userIds[i % userIds.length];
    
    const response = http.post(
      `${BASE_URL}/pullRequest/create`,
      JSON.stringify({
        pull_request_id: prId,
        pull_request_name: `Setup PR ${i}`,
        author_id: authorId,
      }),
      { headers: { 'Content-Type': 'application/json' } }
    );
    
    if (response.status === 201) {
      prIds.push(prId);
    }
  }
  
  console.log(`Test setup complete. Created ${userIds.length} users, ${prIds.length} PRs`);
  return { userIds, prIds, teamNames };
}

export function teardown(data) {
  console.log(`Test completed. Created ${data.prIds.length} PRs for testing.`);
}

export default function(data) {
  const rand = Math.random();
  
  if (rand < 0.30) {
    testCreatePR(data);
  } else if (rand < 0.50) {
    testGetTeam(data);
  } else if (rand < 0.65) {
    testSetIsActive(data);
  } else if (rand < 0.80) {
    testGetReview(data);
  } else if (rand < 0.90) {
    testMergePR(data);
  } else {
    testGetStats();
  }
  
  sleep(0.2 + Math.random() * 0.1);
}

function testCreatePR(data) {
  const prId = `pr_${Date.now()}_${Math.random()}`;
  const authorId = data.userIds[Math.floor(Math.random() * data.userIds.length)];
  
  const response = http.post(
    `${BASE_URL}/pullRequest/create`,
    JSON.stringify({
      pull_request_id: prId,
      pull_request_name: `Test PR ${prId}`,
      author_id: authorId,
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  const success = check(response, {
    'PR created': (r) => r.status === 201,
    'response time < 300ms': (r) => r.timings.duration < 300,
  });
  
  errorRate.add(!success);
}

function testGetTeam(data) {
  const teamName = data.teamNames[Math.floor(Math.random() * data.teamNames.length)];
  
  const response = http.get(`${BASE_URL}/team/get?team_name=${teamName}`);
  
  const success = check(response, {
    'team retrieved': (r) => r.status === 200,
    'response time < 300ms': (r) => r.timings.duration < 300,
  });
  
  errorRate.add(!success);
}

function testSetIsActive(data) {
  const userId = data.userIds[Math.floor(Math.random() * data.userIds.length)];
  
  const response = http.post(
    `${BASE_URL}/users/setIsActive`,
    JSON.stringify({
      user_id: userId,
      is_active: Math.random() > 0.5,
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  const success = check(response, {
    'user updated': (r) => r.status === 200,
    'response time < 300ms': (r) => r.timings.duration < 300,
  });
  
  errorRate.add(!success);
}

function testGetReview(data) {
  const userId = data.userIds[Math.floor(Math.random() * data.userIds.length)];
  
  const response = http.get(`${BASE_URL}/users/getReview?user_id=${userId}`);
  
  const success = check(response, {
    'reviews retrieved': (r) => r.status === 200,
    'response time < 300ms': (r) => r.timings.duration < 300,
  });
  
  errorRate.add(!success);
}

function testMergePR(data) {
  if (data.prIds.length === 0) return;
  
  const prId = data.prIds[Math.floor(Math.random() * data.prIds.length)];
  
  const response = http.post(
    `${BASE_URL}/pullRequest/merge`,
    JSON.stringify({ pull_request_id: prId }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  // 200 = success, 404 = already merged or doesn't exist (acceptable in concurrent test)
  const success = check(response, {
    'merge successful': (r) => r.status === 200 || r.status === 404,
    'response time < 300ms': (r) => r.timings.duration < 300,
  });
  
  errorRate.add(!success);
}



function testGetStats() {
  const response = http.get(`${BASE_URL}/stats`);
  
  const success = check(response, {
    'stats retrieved': (r) => r.status === 200,
    'response time < 300ms': (r) => r.timings.duration < 300,
  });
  
  errorRate.add(!success);
}
