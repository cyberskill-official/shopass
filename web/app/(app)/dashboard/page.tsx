export default function DashboardPage() {
  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Dashboard</h2>
      <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-100">
        <p className="text-gray-600">Welcome to your secure SănDeal dashboard.</p>
        <div className="mt-4 p-4 bg-blue-50 text-blue-800 rounded">
          <p>This is a protected route. If you are seeing this, your session is valid.</p>
        </div>
      </div>
    </div>
  );
}
