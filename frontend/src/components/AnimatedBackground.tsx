export default function AnimatedBackground() {
  return (
    <div className="absolute inset-0">
      {/* Floating Particles */}
      <div className="absolute top-1/4 left-1/4 w-2 h-2 bg-blue-400 rounded-full animate-bounce opacity-60"></div>
      <div className="absolute top-1/3 right-1/3 w-1 h-1 bg-cyan-400 rounded-full animate-pulse opacity-40"></div>
      <div className="absolute bottom-1/3 left-1/3 w-1.5 h-1.5 bg-teal-400 rounded-full animate-bounce delay-1000 opacity-50"></div>
      <div className="absolute bottom-1/4 right-1/4 w-2 h-2 bg-purple-400 rounded-full animate-pulse delay-2000 opacity-60"></div>

      {/* Geometric Shapes */}
      {/* <div className="absolute top-20 left-20 w-8 h-8 border border-blue-400/30 rotate-45 animate-spin"></div> */}
      <div className="absolute top-40 right-20 w-6 h-6 border border-cyan-400/30 rounded-full animate-pulse"></div>
      {/* <div className="absolute bottom-20 left-40 w-4 h-4 bg-gradient-to-r from-teal-400 to-purple-400 rounded-full animate-bounce delay-500"></div> */}
      <div className="absolute bottom-40 left-30 w-5 h-5 border border-purple-400/20 rounded-full animate-pulse delay-1000"></div>

      {/* Gradient Orbs */}
      <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-96 h-96 bg-linear-to-r from-blue-500/10 to-purple-500/10 rounded-full blur-3xl animate-pulse"></div>
      <div className="absolute top-1/3 right-1/3 w-64 h-64 bg-linear-to-r from-cyan-500/10 to-teal-500/10 rounded-full blur-2xl animate-pulse delay-1000"></div>
      <div className="absolute bottom-1/3 left-1/3 w-80 h-80 bg-linear-to-r from-purple-500/10 to-pink-500/10 rounded-full blur-3xl animate-pulse delay-2000"></div>
    </div>
  );
}
